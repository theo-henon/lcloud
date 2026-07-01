package plugin

import (
	"context"
	"net"
	"net/rpc"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/theo-henon/lcloud/internal/volume"
)

type stubHostAPI struct{}

func (stubHostAPI) Emit(_ context.Context, _ Event) error { return nil }
func (stubHostAPI) Log(_ context.Context, _ LogRequest) error { return nil }
func (stubHostAPI) GetVolumeConfig(_ context.Context, volumeID string) (*VolumeConfigView, error) {
	return &VolumeConfigView{ID: volumeID, Name: "demo"}, nil
}
func (stubHostAPI) PluginDataDir(_ context.Context, _, _ string) (string, error) {
	return "/tmp/plugins/demo", nil
}

func TestLcloudPluginRPCServer(t *testing.T) {
	server := &LcloudPluginRPCServer{
		Impl: mockPluginAPI{},
	}
	var resp HandleResult
	require.NoError(t, server.HandleEvent(Event{Type: EventFileUploaded}, &resp))
}

func TestLcloudPluginRPC(t *testing.T) {
	rpc := &LcloudPluginRPC{Impl: mockPluginAPI{}}
	var resp HandleResult
	require.NoError(t, rpc.HandleEvent(Event{Type: EventFileUploaded}, &resp))
}

func TestHostAPIRPCServer(t *testing.T) {
	server := &HostAPIRPCServer{Impl: stubHostAPI{}}
	var empty struct{}
	require.NoError(t, server.Emit(Event{Type: "x"}, &empty))
	require.NoError(t, server.Log(LogRequest{Message: "hi", Level: LogLevelInfo}, &empty))

	var view VolumeConfigView
	require.NoError(t, server.GetVolumeConfig(uuid.NewString(), &view))
	require.Equal(t, "demo", view.Name)

	var path string
	require.NoError(t, server.PluginDataDir(PluginDataDirArgs{VolumeID: "v", PluginID: "p"}, &path))
	require.NotEmpty(t, path)
}

func TestHostAPIRPC(t *testing.T) {
	rpc := &HostAPIRPC{Impl: stubHostAPI{}}
	var empty struct{}
	require.NoError(t, rpc.Emit(Event{Type: "x"}, &empty))
	require.NoError(t, rpc.Log(LogRequest{Message: "hi", Level: LogLevelInfo}, &empty))

	var view VolumeConfigView
	require.NoError(t, rpc.GetVolumeConfig(uuid.NewString(), &view))

	var path string
	require.NoError(t, rpc.PluginDataDir(PluginDataDirArgs{VolumeID: "v", PluginID: "p"}, &path))
}

func TestLcloudPluginPluginServerClient(t *testing.T) {
	pluginImpl := &LcloudPluginPlugin{Impl: mockPluginAPI{}}
	server, err := pluginImpl.Server(nil)
	require.NoError(t, err)
	require.NotNil(t, server)

	hostPlugin := &HostAPIPlugin{Impl: stubHostAPI{}}
	hostServer, err := hostPlugin.Server(nil)
	require.NoError(t, err)
	require.NotNil(t, hostServer)
}

func TestHostAPIGetVolumeConfigDirect(t *testing.T) {
	db := openTestDB(t)
	registry := NewRegistry(db, t.TempDir(), nil)
	volID := uuid.New()
	host := NewHostAPIServer("demo", NewEventBus(), registry, stubVolumeProvider{
		vol: &volume.Volume{ID: volID, Name: "Test"},
	})
	_, err := host.GetVolumeConfig(context.Background(), "not-a-uuid")
	require.Error(t, err)

	view, err := host.GetVolumeConfig(context.Background(), volID.String())
	require.NoError(t, err)
	require.Equal(t, "Test", view.Name)
}

func TestHostAPIRPCClient(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	server := rpc.NewServer()
	require.NoError(t, server.RegisterName("Plugin", &HostAPIRPCServer{Impl: stubHostAPI{}}))
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.ServeConn(conn)
		}
	}()

	client, err := rpc.Dial("tcp", listener.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	api := &HostAPIRPCClient{client: client}
	require.NoError(t, api.Emit(context.Background(), Event{Type: "demo"}))
	require.NoError(t, api.Log(context.Background(), LogRequest{Message: "hello", Level: LogLevelInfo}))

	cfg, err := api.GetVolumeConfig(context.Background(), uuid.NewString())
	require.NoError(t, err)
	require.Equal(t, "demo", cfg.Name)

	path, err := api.PluginDataDir(context.Background(), "vol", "plugin")
	require.NoError(t, err)
	require.NotEmpty(t, path)
}

func TestHostAPIPluginClient(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	server := rpc.NewServer()
	require.NoError(t, server.RegisterName("Plugin", &HostAPIRPCServer{Impl: stubHostAPI{}}))
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.ServeConn(conn)
		}
	}()

	client, err := rpc.Dial("tcp", listener.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	raw, err := (&HostAPIPlugin{}).Client(nil, client)
	require.NoError(t, err)
	host := raw.(HostAPIClient)
	require.NoError(t, host.Emit(context.Background(), Event{Type: "x"}))
}
