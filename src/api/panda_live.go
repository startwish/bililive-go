package api

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"github.com/hr3lxphr6j/bililive-go/src/lib/http"
)

const (
	pandaApiUrl  = "http://www.panda.tv/api_room_v2"
	pandaLiveUrl = "http://pl%s.live.panda.tv/live_panda/%s.flv?sign=%s&ts=%s&rid=%s"
)

type PandaLive struct {
	abstractLive
}

func (p *PandaLive) requestRoomInfo() ([]byte, error) {
	query := map[string]string{
		"roomid": strings.Split(p.Url.Path, "/")[1],
		"__plat": "pc_web",
		"_":      strconv.FormatInt(time.Now().Unix(), 10),
	}
	body, err := http.Get(pandaApiUrl, query, nil)
	if err != nil {
		return nil, err
	}
	if gjson.GetBytes(body, "errno").Int() == 0 {
		return body, nil
	} else {
		return nil, &RoomNotExistsError{p.Url}
	}
}

func (p *PandaLive) GetInfo() (*Info, error) {
	data, err := p.requestRoomInfo()
	if err != nil {
		return nil, err
	}

	info := &Info{
		Live:     p,
		HostName: gjson.GetBytes(data, "data.hostinfo.name").String(),
		RoomName: gjson.GetBytes(data, "data.roominfo.name").String(),
		Status:   gjson.GetBytes(data, "data.videoinfo.status").String() == "2",
	}
	p.cachedInfo = info
	return info, nil
}

func (p *PandaLive) GetStreamUrls() ([]*url.URL, error) {
	data, err := p.requestRoomInfo()
	if err != nil {
		return nil, err
	}
	roomKey := gjson.GetBytes(data, "data.videoinfo.room_key").String()
	plFlag := strings.Split(gjson.GetBytes(data, "data.videoinfo.plflag").String(), "_")
	data2 := gjson.GetBytes(data, "data.videoinfo.plflag_list").String()
	rid := gjson.Get(data2, "auth.rid").String()
	sign := gjson.Get(data2, "auth.sign").String()
	ts := gjson.Get(data2, "auth.time").String()

	u, err := url.Parse(fmt.Sprintf(pandaLiveUrl, plFlag[1], roomKey, sign, ts, rid))
	if err != nil {
		return nil, err
	}
	return []*url.URL{u}, nil
}
