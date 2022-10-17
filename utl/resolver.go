package utl

import (
	"encoding/json"
	"errors"
	etcd "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"golang.org/x/net/context"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
	"log"
)

type EtcdResolverBuilder struct {
	c *etcd.Client
}

func (b *EtcdResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &EtcdResolver{
		target: &target,
		cc:     cc,
		opts:   opts,
		c:      b.c,
		state:  &resolver.State{},
	}
	r.ctx, r.cancel = context.WithCancel(context.TODO())
	if err := r.start(); err != nil {
		return nil, err
	}
	return r, nil
}

func (b *EtcdResolverBuilder) Scheme() string {
	return "testResolver"
}

type EtcdResolver struct {
	target *resolver.Target
	cc     resolver.ClientConn
	opts   resolver.BuildOptions
	c      *etcd.Client
	ctx    context.Context
	cancel context.CancelFunc
	state  *resolver.State
	addrs  map[string]*resolver.Address
}

func (r *EtcdResolver) start() error {
	resp, err := r.c.Get(r.ctx, r.target.URL.Path)
	if err != nil {
		log.Printf("start, getResp fail, err:%v, target.Url:%v", err, r.target.URL)
		return err
	}
	for idx, k := range resp.Kvs {
		v := resp.Kvs[idx].Value
		var update endpoints.Update
		err = json.Unmarshal(v, &update)
		if err != nil {
			log.Printf("unmarshal event:%v, fail, err:%v", v, err)
			continue
		}
		r.addrs[string(k.Key)] = &resolver.Address{
			Addr:               update.Endpoint.Addr,
			ServerName:         "",
			Attributes:         nil,
			BalancerAttributes: nil,
		}
	}
	r.updateState()
	go r.watch(resp.Header.Revision)
	return nil
}

func (r *EtcdResolver) watch(rev int64) {
	watcher := r.c.Watch(r.ctx, r.target.URL.Path, etcd.WithRev(rev))
	for {
		select {
		case <-r.ctx.Done():
			return
		case resp := <-watcher:
			if resp.Err() != nil {
				log.Printf("watcher resp:%v, err:%v", resp, resp.Err())
				return
			}
			for _, evt := range resp.Events {
				var update endpoints.Update
				var err error
				err = json.Unmarshal(evt.Kv.Value, &update)
				if err != nil {
					log.Printf("unmarshal event:%v, fail, err:%v", evt, err)
					continue
				}
				switch evt.Type {
				case etcd.EventTypePut:
					r.addrs[string(evt.Kv.Key)] = &resolver.Address{
						Addr:               update.Endpoint.Addr,
						ServerName:         "",
						Attributes:         nil,
						BalancerAttributes: nil,
					}
				case etcd.EventTypeDelete:
					delete(r.addrs, update.Endpoint.Addr)
				}
			}
			r.updateState()
		}
	}
}

func (r *EtcdResolver) updateState() {
	if len(r.state.Addresses) == 0 {
		return
	}
	r.state.Addresses = nil
	for _, v := range r.addrs {
		r.state.Addresses = append(r.state.Addresses, *v)
	}
	err := r.cc.UpdateState(*r.state)
	if err != nil {
		log.Printf("updateState, err:%v", err)
	}
}

func (r *EtcdResolver) ResolveNow(resolver.ResolveNowOptions) {

}
func (r *EtcdResolver) Close() {
	r.cancel()
	err := r.c.Close()
	if err != nil {
		log.Printf("close etcdClient fail, err:%v", err)
	}
}

const (
	TypEtcd        = "etcd"
	TypPassthrough = "passthrough"
)

// EtcdRegister 只put 不主动delete
type EtcdRegister struct {
	c       *etcd.Client
	ctx     context.Context
	cancel  context.CancelFunc
	leaseId etcd.LeaseID
}

func NewRegister(c *etcd.Client, ttl int64) (*EtcdRegister, error) {
	r := &EtcdRegister{
		c: c,
	}
	r.ctx, r.cancel = context.WithCancel(context.TODO())

	grant, err := r.c.Grant(r.ctx, ttl)
	if err != nil {
		return nil, err
	} else if grant.Error != "" {
		return nil, errors.New(grant.Error)
	}
	keepAlive, err := r.c.KeepAlive(r.ctx, grant.ID)
	if err != nil {
		return nil, err
	}
	go r.keepAlive(keepAlive)
	r.leaseId = grant.ID
	return r, nil
}

func (r *EtcdRegister) register(update endpoints.Update) error {
	data, err := json.Marshal(update)
	if err != nil {
		log.Printf("marshal update:%v fial, err:%v", update, err)
		return err
	}

	resp, err := r.c.Put(context.TODO(), update.Key, string(data), etcd.WithLease(r.leaseId))
	if err != nil {
		return err
	}
	log.Printf("put resp:%v", resp)
	return nil
}

func (r *EtcdRegister) keepAlive(c <-chan *etcd.LeaseKeepAliveResponse) {
	for {
		select {
		case <-r.ctx.Done():
			return
		case resp := <-c:
			log.Printf("keepAliver, resp:%v", resp)
		}
	}
}

func NewBuilder(typ string, client *etcd.Client) (resolver.Builder, error) {
	if typ == TypEtcd {
		return &EtcdResolverBuilder{c: client}, nil
	} else if typ == TypPassthrough {

	}
	return &EtcdResolverBuilder{}, nil
}

type BalancerBuilder struct {
}
type TestBalancer struct {
}

func (b *BalancerBuilder) Build(cc balancer.ClientConn, opts balancer.BuildOptions) balancer.Balancer {
	return nil
}
func (b *BalancerBuilder) Name() string {
	return "testBalancerBuilder"
}

