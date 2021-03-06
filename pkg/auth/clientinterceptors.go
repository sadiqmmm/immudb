/*
Copyright 2019-2020 vChain, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package auth

import (
	"context"

	"google.golang.org/grpc"
)

type WrappedClientStream struct {
	grpc.ClientStream
}

func (w *WrappedClientStream) RecvMsg(m interface{}) error {
	return w.ClientStream.RecvMsg(m)
}

func (w *WrappedClientStream) SendMsg(m interface{}) error {
	return w.ClientStream.SendMsg(m)
}

func ClientStreamInterceptor(token string) func(context.Context, *grpc.StreamDesc, *grpc.ClientConn, string, grpc.Streamer, ...grpc.CallOption) (grpc.ClientStream, error) {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if hasAuth(method) {
			opts = append(opts, grpc.PerRPCCredentials(TokenAuth{
				Token: token,
			}))
		}
		s, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			return nil, err
		}
		return &WrappedClientStream{s}, nil
	}
}

func ClientUnaryInterceptor(token string) func(context.Context, string, interface{}, interface{}, *grpc.ClientConn, grpc.UnaryInvoker, ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if hasAuth(method) {
			opts = append(opts, grpc.PerRPCCredentials(TokenAuth{
				Token: token,
			}))
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

type TokenAuth struct {
	Token string
}

func (t TokenAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.Token,
	}, nil
}

func (TokenAuth) RequireTransportSecurity() bool {
	return false
}
