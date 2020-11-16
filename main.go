package main

import (
	"flag"
	"path"

	"google.golang.org/protobuf/compiler/protogen"
)

var (
	fn, fpkg *string
)

func main() {
	var flags flag.FlagSet
	fn = flags.String("file", "xpc.go", "name of generated go file")
	fpkg = flags.String("pkg", "github.com/celer-network/x-proto-go/xpc", "full go package of generated file")
	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(mygen)
}

func mygen(p *protogen.Plugin) error {
	xpc := p.NewGeneratedFile(path.Join(*fpkg, *fn), protogen.GoImportPath(*fpkg))
	xpc.P("package ", path.Base(*fpkg))
	xpc.P(xpcHdr)
	for _, f := range p.Files {
		if f.Generate {
			for _, svc := range f.Services {
				svcIdt := f.GoImportPath.Ident(svc.GoName)
				for _, rpc := range svc.Methods {
					xpc.P(`case "/`, svcIdt, "/", rpc.GoName, `":`)
					xpc.P(`req = new(`, rpc.Input.GoIdent, ")")
					xpc.P(`resp = new(`, rpc.Output.GoIdent, ")")
				}
			}
		}
	}
	xpc.P(xpcEnd)
	return nil
}

const xpcHdr = `// Copyright 2020 Celer Network

// Code generated by protoc-gen-xpc. DO NOT EDIT.
import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func SendGrpc(con grpc.ClientConnInterface, ctx context.Context, url string, in []byte) ([]byte, error) {
	var req, resp proto.Message
	switch url {
`

const xpcEnd = `
	default:
		return nil, errors.New(url + " not suppoted")
	}
	var err error
	if len(in) > 0 {
		err = protojson.Unmarshal(in, req)
		if err != nil {
			return nil, err
		}
	}
	err = con.Invoke(ctx, url, req, resp)
	if err != nil {
		return nil, err
	}
	jm := protojson.MarshalOptions{
		UseProtoNames: true, // so assign key is same as proto field
	}
	return jm.Marshal(resp)
	}
`
