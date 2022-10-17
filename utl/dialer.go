package utl

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

const (
	dialTimeout = time.Millisecond * 500
)

type Dialer struct {
}

func (d *Dialer) Dial(target string, dialOptions ...grpc.DialOption) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), dialTimeout)
	defer cancel()
	target = "127.0.0.1:5050" // use default addr when resolver is not ready
	conn, err := grpc.DialContext(ctx, target, append(dialOptions, d.defaultDailOptions()...)...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (d *Dialer) defaultCallOptions() []grpc.CallOption {
	return []grpc.CallOption{
		grpc.ForceCodec(&CodeC{}),
	}
}

func (d *Dialer) defaultDailOptions() []grpc.DialOption {
	return []grpc.DialOption{grpc.WithDefaultCallOptions(d.defaultCallOptions()...),
		grpc.WithTransportCredentials(insecure.NewCredentials())}
}

