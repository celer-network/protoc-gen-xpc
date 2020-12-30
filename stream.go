package main

import "html/template"

// for stream template
type oneRpc struct {
	URL, Name, In, Out string
	CS, SS             bool // client stream, server stream
}

var streamSendTpl = template.Must(template.New("send").Parse(`
func StreamSend(con grpc.ClientConnInterface, url string, in []byte) error {
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
	return getStream(con, url, desc).SendMsg(req)
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
		return errors.New(url + " not suppoted")
	}
	s, ok := sMgr.Get(url)
	if !ok {
		return nil, errros.New(url + "try to recv but no stream")
	}
	err := s.RecvMsg(recv)
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
func (s *streamManager) Get(url string) (grpc.ClientStream, bool) {
	m.Lock()
	defer m.Unlock()
	s, ok := s.m[url]
	if s.eof {
		return nil, false
	}
	return s, ok
}
func (s *streamManager) Add(con grpc.ClientConnInterface, url string, desc *grpc.StreamDesc) grpc.ClientStream {
	s, _ := con.NewStream(context.Background(), desc, url)
	m.Lock()
	defer m.Unlock()
	s.m[url] = &oneStream{s, false}
	return s
}

func getStream(con grpc.ClientConnInterface, url string, desc *grpc.StreamDesc) grpc.ClientStream {
	s, ok := sMgr.Get(url)
	if ok {
		return s
	}
	return sMgr.Add(con, url, desc)
}
`
