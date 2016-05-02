package main

import "testing"

func TestGetConfig(t *testing.T) {
	// root := "/zconfig"

	// conn.On("Children", root).Return([]string{})
	_ := map[string]interface{}{}

	// conn.On("Children", root).Return([]string{"servers"})
	// conn.On("Get", root+"/servers").Return("")
	// conn.On("Children", root+"/servers").Return([]string{})
	// /zconfig/servers (nil)
	_ := map[string]interface{}{
		"servers": "",
	}

	// conn.On("Children", root+"/servers").Return([]string{"db"})
	// conn.On("Get", root+"/servers/db").Return("")
	// conn.On("Children", root+"/servers/db").Return([]string{})
	// /zconfig/servers/db (nil)
	_ := map[string]interface{}{
		"servers": []string{"db"},
	}

	// conn.On("Children", root+"/servers").Return([]string{"db", "timeout"})
	// conn.On("Get", root+"/servers/db").Return("")
	// conn.On("Children", root+"/servers/db").Return([]string{})
	// conn.On("Get", root+"/servers/timeout").Return("1000")
	// conn.On("Children", root+"/servers/timeout").Return([]string{})
	// /zconfig/servers/db (nil)
	// /zconfig/servers/timeout (1000)
	_ := map[string]interface{}{
		"servers": map[string]interface{}{
			"db":      "",
			"timeout": "1000",
		},
	}

	// conn.On("Children", root+"/servers/db").Return([]string{"192.168.0.1", "192.168.0.2"})
	// conn.On("Get", root+"/servers/db/192.168.0.1").Return("")
	// conn.On("Get", root+"/servers/db/192.168.0.1").Return("")
	// /zconfig/servers/db/192.168.0.1 (nil)
	// /zconfig/servers/db/192.168.0.2 (nil)
	// /zconfig/servers/timeout (1000)
	_ := map[string]interface{}{
		"servers": map[string]interface{}{
			"db":      []string{"192.168.0.1", "192.168.0.2"},
			"timeout": "1000",
		},
	}
}
