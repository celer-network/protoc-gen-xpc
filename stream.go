package main

import "html/template"

// for stream template
type oneRpc struct {
	URL, Name, In, Out string
	CS, SS             bool // client stream, server stream
}

var streamSendTpl = template.Must(template.New("send").Parse(`
func StreamSend(con grpc.ClientConnInterface, ctx context.Context, url string, in []byte) error {
	var req proto.Message
	desc := new(grpc.StreamDesc)
	switch url {
	{{range .}}
	case "{{.URL}}":
		req = new({{.In}})
		desc.StreamName = "{{.Name}}"
		{{if .SS}}desc.ServerStreams = {{.SS}}{{end}}
		{{if .SS}}desc.ClientStreams = {{.CS}}{{end}}
	{{end}}
	default:
		return errors.New(url + " not suppoted")
	}
	var err error
	if len(in) > 0 {
		err = protojson.Unmarshal(in, req)
		if err != nil {
			return err
		}
	}
	s, ok := sMgr.Get(url)
	if ok {
		return s.SendMsg(req)
	}
	return sMgr.Add(con, ctx, url, desc).SendMsg(req)
}`))

var streamRecvTpl = template.Must(template.New("recv").Parse(`
func StreamRecv(url string) ([]byte, error) {
	var recv proto.Message
	switch url {
	{{range .}}
	case "{{.URL}}":
		recv = new({{.Out}})
	{{end}}
	default:
		return nil, errors.New(url + " not suppoted")
	}
	s, ok := sMgr.Get(url)
	if !ok {
		return nil, errors.New(url + "try to recv but no stream")
	}
	err := s.RecvMsg(recv)
	if err == io.EOF {
		s.eof = true
	}
	if err != nil {
		return nil, err
	}
	return jm.Marshal(recv)
}`))

// common code for streaming rpcs
// getStream will create or reuse
const streamUtil = `
// below are stream utils/helper
var sMgr = &streamManager{
	m: make(map[string]*oneStream),
}
type oneStream struct {
	grpc.ClientStream
	eof bool
}
type streamManager struct {
	m map[string]*oneStream
	sync.Mutex
}
func (sm *streamManager) Get(url string) (*oneStream, bool) {
	sm.Lock()
	defer sm.Unlock()
	cs, ok := sm.m[url]
    if cs == nil {
		return nil, false
	}
	if cs.eof {
		return nil, false
	}
	return cs, ok
}
func (sm *streamManager) Add(con grpc.ClientConnInterface, ctx context.Context, url string, desc *grpc.StreamDesc) *oneStream {
	cs, _ := con.NewStream(ctx, desc, url)
	sm.Lock()
	defer sm.Unlock()
	news := &oneStream{cs, false}
	sm.m[url] = news
	return news
}
`
