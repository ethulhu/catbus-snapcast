package snapcast

// https://github.com/badaix/snapcast/blob/master/doc/json_rpc_api/v2_0_0.md.

type (
	// common structs.

	host struct {
		IP   string `json:"ip,omitempty"`
		Name string `json:"name"`
		Arch string `json:"arch,omitempty"`
		OS   string `json:"os,omitempty"`
		MAC  string `json:"mac,omitempty"`
	}

	volume struct {
		Muted   bool `json:"muted"`
		Percent int  `json:"percent"`
	}

	clientStatus struct {
		ID        string `json:"id"`
		Connected bool   `json:"connected"`
		Host      host   `json:"host"`
		Config    struct {
			Name     string `json:"name"`
			Instance int    `json:"instance"`
			Latency  int    `json:"latency"`
			Volume   volume `json:"volume"`
		} `json:"config"`
		Snapclient struct {
			ProtocolVersion int    `json:"protocolVersion"`
			Version         string `json:"version"`
			Name            string `json:"name"`
		} `json:"snapclient"`
		LastSeen struct {
			Sec  int `json:"sec"`
			Usec int `json:"usec"`
		} `json:"lastSeen"`
	}

	groupStatus struct {
		ID      string         `json:"id"`
		Name    string         `json:"name"`
		Muted   bool           `json:"muted"`
		Stream  Stream         `json:"stream_id"`
		Clients []clientStatus `json:"clients"`
	}

	streamStatus struct {
		ID     string            `json:"id"`
		Status string            `json:"status"`
		Meta   map[string]string `json:"meta"`
		URI    struct {
			Host     string `json:"host"`
			Fragment string `json:"fragment"`
			Query    struct {
				Codec              string `json:"codec"`
				BufferMilliseconds string `json:"buffer_ms"`
				Name               string `json:"name"`
				SampleFormat       string `json:"sampleformat"`
			} `json:"query"`
			Scheme string `json:"scheme"`
			Raw    string `json:"raw"`
			Path   string `json:"path"`
		} `json:"uri"`
	}

	serverStatus struct {
		Snapserver struct {
			ProtocolVersion        int    `json:"protocolVersion"`
			Version                string `json:"version"`
			Name                   string `json:"name"`
			ControlProtocolVersion int    `json:"controlProtocolVersion"`
		} `json:"snapserver"`
		Host host `json:"host"`
	}

	// RPC requests & responses.

	serverGetRPCVersionResponse struct {
		Major int `json:"major"`
		Minor int `json:"minor"`
		Patch int `json:"patch"`
	}

	serverGetStatusResponse struct {
		Server struct {
			Streams []streamStatus `json:"streams"`
			Groups  []groupStatus  `json:"groups"`
			Server  serverStatus   `json:"server"`
		} `json:"server"`
	}

	clientGetStatusRequest struct {
		ID string `json:"id"`
	}
	clientGetStatusResponse struct {
		Client clientStatus `json:"client"`
	}

	clientSetVolumeRequest struct {
		ID     string `json:"id"`
		Volume volume `json:"volume"`
	}
	clientSetVolumeResponse struct {
		Volume volume `json:"volume"`
	}

	clientSetNameRequest struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	clientSetNameResponse struct {
		Name string `json:"name"`
	}

	groupGetStatusRequest struct {
		ID string `json:"id"`
	}
	groupGetStatusResponse struct {
		Group groupStatus `json:"group"`
	}

	groupSetStreamRequest struct {
		ID     string `json:"id"`
		Stream Stream `json:"stream_id"`
	}
	groupSetStreamResponse struct {
		Stream Stream `json:"stream_id"`
	}

	groupSetNameRequest struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	groupSetNameResponse struct {
		Name string `json:"name"`
	}
)

// JSON-RPC method names.
const (
	clientGetStatus  = "Client.GetStatus"
	clientSetLatency = "Client.SetLatency"
	clientSetName    = "Client.SetName"
	clientSetVolume  = "Client.SetVolume"

	groupGetStatus  = "Group.GetStatus"
	groupSetClients = "Group.SetClients"
	groupSetMute    = "Group.SetMute"
	groupSetName    = "Group.SetName"
	groupSetStream  = "Group.SetStream"

	serverGetRPCVersion = "Server.GetRPCVersion"
	serverGetStatus     = "Server.GetStatus"
	serverDeleteClient  = "Server.DeleteClient"

	streamAddStream    = "Stream.AddStream"
	streamRemoveStream = "Stream.RemoveStream"
)
