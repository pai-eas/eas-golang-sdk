package eas

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/xd-luqiang/eas-golang-sdk/eas/types"
	"golang.org/x/net/websocket"
)

const (
	HeaderRequestId     = "X-Eas-Queueservice-Request-Id"
	HeaderAccessRear    = "X-EAS-QueueService-Access-Rear"
	HeaderAuthorization = "Authorization"

	DefaultBasePath = "/api/predict"

	DefaultGroupName = "eas"
)

// testing only.
type breakGenerator func() bool

func noBreak() bool {
	return false
}

type QueueUser struct {
	uid   string
	gid   string
	token string
}

func NewQueueUser(uid, gid, token string) QueueUser {
	return QueueUser{
		uid, gid, token,
	}
}

func (c QueueUser) Uid() string {
	return c.uid
}

func (c QueueUser) Gid() string {
	return c.gid
}

func (c QueueUser) Token() string {
	return c.token
}

// QueueClient is client of queue server, which also implements queue service interface
type QueueClient struct {
	// HTTP client.
	httpClient *http.Client
	// base url of queue server.
	baseUrl *url.URL

	extraHeader map[string]string

	user types.User

	WebsocketWatch bool
	bg             breakGenerator

	once sync.Once
	attr types.Attributes
	// codecs for data frame and attributes.
	DCodec types.DataFrameCodec
	ACodec types.AttributesCodec
}

type queueOptions struct {
	extraHeaders map[string]string
	basePath     string
	uid          string
	gid          string
	bg           breakGenerator
}

type QueueOption func(*queueOptions)

func WithExtraHeaders(extraHeaders map[string]string) QueueOption {
	return func(o *queueOptions) {
		o.extraHeaders = extraHeaders
	}
}

func WithBasePath(basePath string) QueueOption {
	return func(o *queueOptions) {
		o.basePath = basePath
	}
}

func WithUserId(uid string) QueueOption {
	return func(o *queueOptions) {
		o.uid = uid
	}
}

func WithGroupId(gid string) QueueOption {
	return func(o *queueOptions) {
		o.gid = gid
	}
}

func withBreakGenerator(bg breakGenerator) QueueOption {
	return func(o *queueOptions) {
		o.bg = bg
	}
}

func NewQueueClient(endpoint, queueName, token string, opts ...QueueOption) (*QueueClient, error) {
	queueOpt := &queueOptions{basePath: DefaultBasePath, bg: noBreak}
	for _, opt := range opts {
		opt(queueOpt)
	}
	baseUrl := endpoint + path.Join("/", queueOpt.basePath, queueName)
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	if len(u.Scheme) == 0 {
		u.Scheme = "http"
	}
	if len(queueOpt.uid) == 0 {
		queueOpt.uid = uuid.New().String()
	}
	if len(queueOpt.gid) == 0 {
		queueOpt.gid = DefaultGroupName
	}
	cli := &QueueClient{
		baseUrl:        u,
		bg:             queueOpt.bg,
		httpClient:     &http.Client{},
		user:           NewQueueUser(queueOpt.uid, queueOpt.gid, token),
		WebsocketWatch: true, // Watch through websocket by default
		extraHeader:    queueOpt.extraHeaders,
		DCodec:         types.DataFrameCodecFor(types.ContentTypeProtobuf),
		ACodec:         types.AttributesCodecFor(types.ContentTypeProtobuf),
	}

	return cli, nil
}

func readMessage(reader io.Reader) string {
	b, err := io.ReadAll(reader)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func (q *QueueClient) getAttr(force bool) (types.Attributes, error) {
	var err error
	defer func() {
		if err != nil && len(q.attr) == 0 {
			q.reset()
		}
	}()
	if len(q.attr) == 0 {
		q.once.Do(func() { err = q.obtainAttr() })
	} else if force {
		if err = q.obtainAttr(); err != nil {
			return q.attr, fmt.Errorf("failed to obtain attributes, error: %v", err)
		}
	}

	return q.attr, err
}

func (q *QueueClient) reset() {
	q.attr = nil
	q.once = sync.Once{}
}

func (q *QueueClient) obtainAttr() error {
	// make a copy of base url.
	u := *q.baseUrl
	qe := u.Query()
	qe.Set("_attrs_", "true")
	u.RawQuery = qe.Encode()
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	q.withAuthorization(req)
	q.AddExtraHeaders(req.Header)
	req.Header.Set("accept", q.ACodec.MediaType())
	resp, err := q.httpClient.Do(req)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("visiting: %s, unexpected status code: %d, body: %s", u.String(), resp.StatusCode, string(body))
	}
	attr := types.Attributes{}
	if err = q.ACodec.Decode(body, &attr); err != nil {
		return err
	}
	q.attr = attr
	return nil
}

// withIdentity populates user and group id into request.
func (q *QueueClient) withIdentity(req *http.Request) error {
	attr, err := q.getAttr(false)
	if err != nil {
		return err
	}
	uidHeader := attr[types.UserIdentifyHeader]
	gidHeader := attr[types.GroupIdentifyHeader]
	if len(uidHeader) == 0 {
		return fmt.Errorf("malformed attributes: %v", attr)
	} else {
		req.Header.Add(uidHeader, q.user.Uid())
	}
	if len(gidHeader) > 0 {
		req.Header.Add(gidHeader, q.user.Gid())
	}
	return nil
}

// withIdentity populates user and group id into request.
func (q *QueueClient) withAuthorization(req *http.Request) {
	if t, ok := q.user.(types.UserWithToken); ok {
		req.Header.Add(HeaderAuthorization, t.Token())
	}
}

func (q *QueueClient) withPriority(req *http.Request, prio types.Priority) error {
	if prio > 0 {
		attr, err := q.getAttr(false)
		if err != nil {
			return err
		}
		ph, ok := attr[types.PriorityHeader]
		if !ok {
			return nil
		}
		req.Header.Add(ph, strconv.FormatInt(int64(prio), 10))
	}
	return nil
}

func (q *QueueClient) AddExtraHeaders(header http.Header) {
	for key, val := range q.extraHeader {
		header.Set(key, val)
	}
}

// Truncate truncates the queue to the given index, the specified index is not included.
func (q *QueueClient) Truncate(ctx context.Context, index uint64) error {
	// make a copy of base url.
	u := *q.baseUrl
	eq := u.Query()
	eq.Set("_index_", strconv.FormatUint(index, 10))
	eq.Set("_trunc_", boolString(true))
	u.RawQuery = eq.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u.String(), nil)
	if err != nil {
		return err
	}
	if err := q.withIdentity(req); err != nil {
		return err
	}
	q.withAuthorization(req)
	q.AddExtraHeaders(req.Header)
	resp, err := q.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("visiting: %s, unexpected status code: %d, message: %s", u.String(), resp.StatusCode, readMessage(resp.Body))
	}
	return nil
}

func (q *QueueClient) End(ctx context.Context, force bool) error {
	// make a copy of base url.
	u := *q.baseUrl
	eq := u.Query()
	eq.Set("_eos_", boolString(true))
	if force {
		eq.Set("_force_", boolString(true))
	}
	u.RawQuery = eq.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		return err
	}
	if err := q.withIdentity(req); err != nil {
		return err
	}
	q.withAuthorization(req)
	q.AddExtraHeaders(req.Header)
	resp, err := q.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("visiting: %s, unexpected status code: %d, message: %s", u.String(), resp.StatusCode, readMessage(resp.Body))
	}
	return nil
}

// Put puts data into queue. It returns the index of the data in queue, and generated request id.
func (q *QueueClient) Put(ctx context.Context, data []byte, tags types.Tags) (uint64, string, error) {
	return q.PutWithPathAndPriority(ctx, data, "", tags, 0)
}

func (q *QueueClient) PutWithPath(ctx context.Context, data []byte, path string, tags types.Tags) (uint64, string, error) {
	return q.PutWithPathAndPriority(ctx, data, path, tags, 0)
}

func (q *QueueClient) PutWithPriority(ctx context.Context, data []byte, tags types.Tags, prio types.Priority) (index uint64, requestId string, err error) {
	return q.PutWithPathAndPriority(ctx, data, "", tags, prio)
}

// PutWithPriority puts data into queue with path and priority.
// It returns the index of the data in queue, and generated request id.
// The prioritized data will be received by Watcher before normal data.
func (q *QueueClient) PutWithPathAndPriority(ctx context.Context, data []byte, customPath string, tags types.Tags, prio types.Priority) (index uint64, requestId string, err error) {
	// make a copy of base url.
	u := *q.baseUrl
	qe := u.Query()
	for key, val := range tags {
		qe.Set(key, val)
	}
	u.RawQuery = qe.Encode()
	if customPath != "" {
		u.Path = path.Join(u.Path, customPath)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(data))
	if err != nil {
		return 0, requestId, err
	}
	if err := q.withIdentity(req); err != nil {
		return 0, requestId, err
	}
	q.withAuthorization(req)
	if err := q.withPriority(req, prio); err != nil {
		return 0, requestId, err
	}
	q.AddExtraHeaders(req.Header)
	resp, err := q.httpClient.Do(req)
	if err != nil {
		return 0, requestId, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, requestId, err
	}
	requestId = resp.Header.Get(HeaderRequestId)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return 0, requestId, fmt.Errorf("visiting: %s, unexpected status code: %d, message: %s", u.String(), resp.StatusCode, string(body))
	}
	defer resp.Body.Close()
	index, err = strconv.ParseUint(string(body), 0, 64)
	if err != nil {
		return 0, requestId, err
	}
	return index, requestId, nil
}

// GetByIndex gets data from queue by index,  make convenience wrapper for Get.
func (q *QueueClient) GetByIndex(ctx context.Context, index uint64) (dfs []types.DataFrame, err error) {
	return q.Get(ctx, index, 1, time.Duration(0), true, types.Tags{})
}

// GetByRequestId gets data from queue by request id,  make convenience wrapper for Get.
func (q *QueueClient) GetByRequestId(ctx context.Context, requestId string) (dfs []types.DataFrame, err error) {
	return q.Get(ctx, 0, 1, time.Duration(0), true, types.Tags{"requestId": requestId})
}

// Get gets data from queue, it returns the data frames and error.
// Parameters:
//   - index: the start point to get data, if index is 0, it will search from the queue head.
//   - length: the number of data frames to get.
//   - timeout: the timeout duration to wait for data, if timeout is 0, it will return immediately.
//   - autoDelete: if autoDelete is true, the data will be deleted from queue after it is read.
//   - tags: the tags to filter data.
func (q *QueueClient) Get(ctx context.Context, index uint64, length int, timeout time.Duration, autoDelete bool, tags types.Tags) (dfs []types.DataFrame, err error) {
	var ret []types.DataFrame
	u := *q.baseUrl
	eq := u.Query()
	eq.Set("_index_", strconv.FormatUint(index, 10))
	eq.Set("_length_", strconv.FormatInt(int64(length), 10))
	eq.Set("_timeout_", timeout.String())
	eq.Set("_raw_", boolString(false))
	eq.Set("_auto_delete_", boolString(autoDelete))
	if err = tags.Validate(); err != nil {
		return nil, err
	}
	for key, val := range tags {
		eq.Set(key, val)
	}
	u.RawQuery = eq.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return ret, err
	}
	req.Header.Set("Accept", q.DCodec.MediaType())
	if err := q.withIdentity(req); err != nil {
		return ret, err
	}
	q.withAuthorization(req)
	q.AddExtraHeaders(req.Header)
	resp, err := q.httpClient.Do(req)
	if err != nil {
		return ret, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return ret, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return ret, fmt.Errorf("visiting: %s, unexpected status code: %d, message: %s", u.String(), resp.StatusCode, string(data))
	}

	return q.DCodec.DecodeList(data)
}

func boolString(b bool) string {
	if b {
		return "true"
	} else {
		return "false"
	}
}

type websocketWatcher struct {
	ctx             context.Context
	conn            *websocket.Conn
	decoder         types.DataFrameDecoder
	pingFrameWriter io.WriteCloser
	ch              chan types.DataFrame
	bg              breakGenerator

	closed int32
}

func newWebsocketWatcher(ctx context.Context, config *websocket.Config, decoder types.DataFrameDecoder) (types.Watcher, error) {
	conn, err := websocket.DialConfig(config)
	if err != nil {
		return nil, err
	}
	ping, err := conn.NewFrameWriter(websocket.PingFrame)
	if err != nil {
		return nil, err
	}
	w := &websocketWatcher{
		ctx:             ctx,
		conn:            conn,
		decoder:         decoder,
		pingFrameWriter: ping,
		ch:              make(chan types.DataFrame, 100),
	}
	go w.run()
	return w, nil
}

func (w *websocketWatcher) FrameChan() <-chan types.DataFrame {
	return w.ch
}

func (w *websocketWatcher) Close() {
	if atomic.CompareAndSwapInt32(&w.closed, 0, 1) {
		_ = w.conn.Close()
		close(w.ch)
	}
}

func (w *websocketWatcher) pingServer() error {
	_, err := w.pingFrameWriter.Write([]byte{})
	return err
}

func (w *websocketWatcher) push(df types.DataFrame) {
	defer noPanic()
	if atomic.LoadInt32(&w.closed) == 0 {
		w.ch <- df
	}
}

func (w *websocketWatcher) run() {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		defer w.pingFrameWriter.Close()
		for atomic.LoadInt32(&w.closed) == 0 {
			select {
			case <-w.ctx.Done():
				_ = w.conn.Close()
				return
			case <-ticker.C:
				_ = w.pingServer()
			}
		}
	}()
	defer w.Close()
	var data []byte
	for {
		df := types.DataFrame{}
		err := websocket.Message.Receive(w.conn, &data)
		if err != nil {
			df.Message = fmt.Sprintf("error reading message: %v", err)
			w.push(df)
			return
		}
		if err = w.decoder.Decode(data, &df); err != nil {
			df.Message = fmt.Sprintf("failed to decode message: %v", err)
			w.push(df)
			return
		}
		w.push(df)
	}
}

type watcherCtor func(context.Context, types.DataFrameDecoder) (types.Watcher, error)

type reconnectWatcher struct {
	delegator types.Watcher
	userChan  chan types.DataFrame
	ctx       context.Context
	cancel    context.CancelFunc
	decoder   types.DataFrameDecoder
	wCtor     watcherCtor

	closed int32
	// only for testing
	bg breakGenerator
}

func (w *reconnectWatcher) push(df types.DataFrame) {
	defer noPanic()
	if atomic.LoadInt32(&w.closed) == 0 {
		w.userChan <- df
	}
}

func newReconnectWatcher(ctx context.Context, wCtor watcherCtor, decoder types.DataFrameDecoder, bg breakGenerator) (types.Watcher, error) {
	wCtx, wCancel := context.WithCancel(ctx)
	watcher, err := wCtor(wCtx, decoder)
	if err != nil {
		wCancel()
		return nil, err
	}
	w := &reconnectWatcher{
		delegator: watcher,
		userChan:  make(chan types.DataFrame, 100),
		decoder:   decoder,
		ctx:       wCtx,
		cancel:    wCancel,
		wCtor:     wCtor,
		bg:        bg,
	}
	go w.run()
	return w, nil
}

func (w *reconnectWatcher) FrameChan() <-chan types.DataFrame {
	return w.userChan
}

func (w *reconnectWatcher) Close() {
	if atomic.CompareAndSwapInt32(&w.closed, 0, 1) {
		w.cancel()
		w.delegator.Close()
		close(w.userChan)
	}
}

func (w *reconnectWatcher) run() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	defer w.Close()
	for {
		if w.bg() {
			w.delegator.Close()
		}
		df, ok := <-w.delegator.FrameChan()
		// connection closed
		if !ok {
			fmt.Println("Upstream connection closed, reconnecting...")
			// connection was closed by upstream unexpectedly, try to reconnect
		loop:
			for {
				select {
				case <-ticker.C:
					// try to reconnect every 1 second
					watcher, err := w.wCtor(w.ctx, w.decoder)
					if err != nil {
						fmt.Printf("Connect to upstream error: %v, retry...\n", err)
						continue
					}
					w.delegator.Close()
					w.delegator = watcher
					break loop
				case <-w.ctx.Done():
					// watcher was closed by user
					return
				}
			}
		} else {
			w.push(df)
		}
	}
}

type httpWatcher struct {
	ctx     context.Context
	reader  io.ReadCloser
	decoder types.DataFrameDecoder
	ch      chan types.DataFrame

	closed int32
}

func newHTTPWatcher(ctx context.Context, reader io.ReadCloser, decoder types.DataFrameDecoder) *httpWatcher {
	w := &httpWatcher{
		ctx:     ctx,
		reader:  reader,
		decoder: decoder,
		ch:      make(chan types.DataFrame, 100),
	}
	go w.run()
	return w
}

func (h *httpWatcher) FrameChan() <-chan types.DataFrame {
	return h.ch
}

func (h *httpWatcher) Close() {
	if atomic.CompareAndSwapInt32(&h.closed, 0, 1) {
		_ = h.reader.Close()
		close(h.ch)
	}
}

func (h *httpWatcher) push(df types.DataFrame) {
	defer noPanic()
	if atomic.LoadInt32(&h.closed) == 0 {
		h.ch <- df
	}
}

func (h *httpWatcher) run() {
	defer h.Close()
	rbuf := [4096]byte{}
	buf := bytes.NewBuffer(nil)
	for {
		n, err := h.reader.Read(rbuf[:])
		if n > 0 {
			io.Copy(buf, bytes.NewBuffer(rbuf[:n]))
			if errors.Is(err, io.ErrShortBuffer) {
				continue
			} else if err != nil {
				// fmt.Printf("failed to read: %v\n", err)
				return
			}
			df := types.DataFrame{}
			if err = h.decoder.Decode(buf.Bytes(), &df); err != nil {
				// fmt.Errorf("failed to decode, err: %v", err)
				return
			}
			buf.Reset()
			h.push(df)
		} else {
			break
		}
	}
}

func (q *QueueClient) Watch(ctx context.Context, index, window uint64, indexOnly bool, autocommit bool) (types.Watcher, error) {
	return q.WatchByTag(ctx, index, window, indexOnly, autocommit, types.Tags{})
}

// WatchByTag returns a Watcher which can be used to receive data from the queue in streaming fashion.
// Parameters:
//   - index: the index to start watching from.
//   - window: the window size to watch.
//   - indexOnly: if true, only the index will be returned, otherwise the whole data frame will be returned.
//   - autocommit: if true, the index will be automatically committed after the data frame is received.
//   - tags: the tags to watch.
func (q *QueueClient) WatchByTag(ctx context.Context, index, window uint64, indexOnly bool, autocommit bool, tags types.Tags) (types.Watcher, error) {
	u := *q.baseUrl
	eq := u.Query()
	eq.Set("_index_", strconv.FormatUint(index, 10))
	eq.Set("_window_", strconv.FormatUint(window, 10))
	eq.Set("_index_only_", boolString(indexOnly))
	eq.Set("_auto_commit_", boolString(autocommit))
	eq.Set("_watch_", boolString(true))
	if err := tags.Validate(); err != nil {
		return nil, err
	}
	for key, val := range tags {
		eq.Set(key, val)
	}
	u.RawQuery = eq.Encode()
	if q.WebsocketWatch {
		// use websocket watch.
		u.Scheme = "ws"
		config, err := websocket.NewConfig(u.String(), q.baseUrl.String())
		if err != nil {
			return nil, err
		}
		header := http.Header{}
		attr, err := q.getAttr(true)
		if err != nil {
			return nil, err
		}
		uidHeader := attr[types.UserIdentifyHeader]
		gidHeader := attr[types.GroupIdentifyHeader]
		// set websocket request headers.
		header.Set(uidHeader, q.user.Uid())
		header.Set("Accept", q.DCodec.MediaType())
		header.Set(HeaderAuthorization, q.user.Token())
		if len(gidHeader) > 0 {
			header.Set(gidHeader, q.user.Gid())
		}
		config.Header = header
		ctor := func(ctx context.Context, decoder types.DataFrameDecoder) (types.Watcher, error) {
			return newWebsocketWatcher(ctx, config, decoder)
		}
		return newReconnectWatcher(ctx, ctor, q.DCodec, q.bg)

	} else {
		// http watch.
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", q.DCodec.MediaType())
		if err := q.withIdentity(req); err != nil {
			return nil, err
		}
		q.withAuthorization(req)
		q.AddExtraHeaders(req.Header)
		ctor := func(ctx context.Context, decoder types.DataFrameDecoder) (types.Watcher, error) {
			resp, err := q.httpClient.Do(req)
			if err != nil {
				return nil, err
			}
			if resp.StatusCode != 200 {
				content, _ := io.ReadAll(resp.Body)
				return nil, fmt.Errorf("unexpected status code: %d, message: %s", resp.StatusCode, string(content))
			}
			reader := types.NewLengthDelimitedFrameReader(resp.Body)
			return newHTTPWatcher(ctx, reader, decoder), nil
		}
		return newReconnectWatcher(ctx, ctor, q.DCodec, q.bg)
	}
}

// Commit commits the indexes to the queue, as the result, the data in queue will not be delivered again.
func (q *QueueClient) Commit(ctx context.Context, indexes ...uint64) error {
	// make a copy of base url.
	u := *q.baseUrl
	var indexStr []string
	for _, idx := range indexes {
		indexStr = append(indexStr, strconv.FormatUint(idx, 10))
	}
	eq := u.Query()
	eq.Set("_indexes_", strings.Join(indexStr, ","))
	// eq.Set("_delete_", boolString(del))
	u.RawQuery = eq.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), nil)
	if err != nil {
		return err
	}
	if err := q.withIdentity(req); err != nil {
		return err
	}
	q.withAuthorization(req)
	q.AddExtraHeaders(req.Header)
	resp, err := q.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("visiting: %s, unexpected status code %d, message: %s", u.String(), resp.StatusCode, readMessage(resp.Body))
	}
	return nil
}

func (q *QueueClient) Negative(ctx context.Context, code types.Code, reason string, indexes ...uint64) error {
	// make a copy of base url.
	u := *q.baseUrl
	var indexStr []string
	for _, idx := range indexes {
		indexStr = append(indexStr, strconv.FormatUint(idx, 10))
	}
	eq := u.Query()
	eq.Set("_indexes_", strings.Join(indexStr, ","))
	eq.Set("_negative_", strconv.FormatBool(true))
	u.RawQuery = eq.Encode()
	formData := url.Values{"_reason_": {reason}, "_code_": {code.String()}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := q.withIdentity(req); err != nil {
		return err
	}
	q.withAuthorization(req)
	q.AddExtraHeaders(req.Header)
	resp, err := q.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("visiting: %s, unexpected status code %d, message: %s", u.String(), resp.StatusCode, readMessage(resp.Body))
	}
	return nil
}

// Del deletes the indexes from the queue, the content of the indexes will also be deleted.
func (q *QueueClient) Del(ctx context.Context, indexes ...uint64) error {
	// make a copy of base url.
	u := *q.baseUrl
	var indexStr []string
	for _, idx := range indexes {
		indexStr = append(indexStr, strconv.FormatUint(idx, 10))
	}
	eq := u.Query()
	eq.Set("_indexes_", strings.Join(indexStr, ","))
	u.RawQuery = eq.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u.String(), nil)
	if err != nil {
		return err
	}
	if err := q.withIdentity(req); err != nil {
		return err
	}
	q.withAuthorization(req)
	q.AddExtraHeaders(req.Header)
	resp, err := q.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("visiting: %s, unexpected status code %d, message: %s", u.String(), resp.StatusCode, readMessage(resp.Body))
	}
	return nil
}

func (q *QueueClient) Attributes() (types.Attributes, error) {
	return q.getAttr(true)
}
