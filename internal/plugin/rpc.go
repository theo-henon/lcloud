package plugin

import (
	"context"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

const PluginName = "lcloud"

type LcloudPlugin interface {
	HandleEvent(ctx context.Context, event Event) (*HandleResult, error)
}

type HostAPIClient interface {
	Emit(ctx context.Context, event Event) error
	Log(ctx context.Context, req LogRequest) error
	GetVolumeConfig(ctx context.Context, volumeID string) (*VolumeConfigView, error)
	PluginDataDir(ctx context.Context, volumeID, pluginID string) (string, error)
}

type LcloudPluginRPC struct {
	Impl LcloudPlugin
}

func (r *LcloudPluginRPC) HandleEvent(args Event, resp *HandleResult) error {
	result, err := r.Impl.HandleEvent(context.Background(), args)
	if err != nil {
		return err
	}
	if result == nil {
		*resp = HandleResult{}
		return nil
	}
	*resp = *result
	return nil
}

type LcloudPluginRPCServer struct {
	Impl LcloudPlugin
}

func (s *LcloudPluginRPCServer) HandleEvent(args Event, resp *HandleResult) error {
	result, err := s.Impl.HandleEvent(context.Background(), args)
	if err != nil {
		return err
	}
	if result == nil {
		*resp = HandleResult{}
		return nil
	}
	*resp = *result
	return nil
}

type LcloudPluginPlugin struct {
	Impl LcloudPlugin
}

func (p *LcloudPluginPlugin) Server(*plugin.MuxBroker) (any, error) {
	return &LcloudPluginRPCServer{Impl: p.Impl}, nil
}

func (LcloudPluginPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (any, error) {
	return &LcloudPluginRPCClient{client: c}, nil
}

type LcloudPluginRPCClient struct {
	client *rpc.Client
}

func (c *LcloudPluginRPCClient) HandleEvent(ctx context.Context, event Event) (*HandleResult, error) {
	var resp HandleResult
	err := c.client.Call("Plugin.HandleEvent", event, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type HostAPIRPC struct {
	Impl HostAPIClient
}

func (r *HostAPIRPC) Emit(args Event, resp *struct{}) error {
	return r.Impl.Emit(context.Background(), args)
}

func (r *HostAPIRPC) Log(args LogRequest, resp *struct{}) error {
	return r.Impl.Log(context.Background(), args)
}

func (r *HostAPIRPC) GetVolumeConfig(args string, resp *VolumeConfigView) error {
	cfg, err := r.Impl.GetVolumeConfig(context.Background(), args)
	if err != nil {
		return err
	}
	*resp = *cfg
	return nil
}

func (r *HostAPIRPC) PluginDataDir(args PluginDataDirArgs, resp *string) error {
	path, err := r.Impl.PluginDataDir(context.Background(), args.VolumeID, args.PluginID)
	if err != nil {
		return err
	}
	*resp = path
	return nil
}

type PluginDataDirArgs struct {
	VolumeID string
	PluginID string
}

type HostAPIRPCServer struct {
	Impl HostAPIClient
}

func (s *HostAPIRPCServer) Emit(args Event, resp *struct{}) error {
	return s.Impl.Emit(context.Background(), args)
}

func (s *HostAPIRPCServer) Log(args LogRequest, resp *struct{}) error {
	return s.Impl.Log(context.Background(), args)
}

func (s *HostAPIRPCServer) GetVolumeConfig(args string, resp *VolumeConfigView) error {
	cfg, err := s.Impl.GetVolumeConfig(context.Background(), args)
	if err != nil {
		return err
	}
	*resp = *cfg
	return nil
}

func (s *HostAPIRPCServer) PluginDataDir(args PluginDataDirArgs, resp *string) error {
	path, err := s.Impl.PluginDataDir(context.Background(), args.VolumeID, args.PluginID)
	if err != nil {
		return err
	}
	*resp = path
	return nil
}

type HostAPIPlugin struct {
	Impl HostAPIClient
}

func (p *HostAPIPlugin) Server(b *plugin.MuxBroker) (any, error) {
	return &HostAPIRPCServer{Impl: p.Impl}, nil
}

func (HostAPIPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (any, error) {
	return &HostAPIRPCClient{client: c}, nil
}

type HostAPIRPCClient struct {
	client *rpc.Client
}

func (c *HostAPIRPCClient) Emit(ctx context.Context, event Event) error {
	var resp struct{}
	return c.client.Call("Plugin.Emit", event, &resp)
}

func (c *HostAPIRPCClient) Log(ctx context.Context, req LogRequest) error {
	var resp struct{}
	return c.client.Call("Plugin.Log", req, &resp)
}

func (c *HostAPIRPCClient) GetVolumeConfig(ctx context.Context, volumeID string) (*VolumeConfigView, error) {
	var resp VolumeConfigView
	err := c.client.Call("Plugin.GetVolumeConfig", volumeID, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *HostAPIRPCClient) PluginDataDir(ctx context.Context, volumeID, pluginID string) (string, error) {
	var resp string
	err := c.client.Call("Plugin.PluginDataDir", PluginDataDirArgs{VolumeID: volumeID, PluginID: pluginID}, &resp)
	if err != nil {
		return "", err
	}
	return resp, nil
}

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "LCLOUD_PLUGIN",
	MagicCookieValue: "lcloud-plugin-v1",
}

var _ LcloudPlugin = (*LcloudPluginRPCClient)(nil)
