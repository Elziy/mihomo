package statistic

import "github.com/metacubex/mihomo/common/atomic"

type RuleStatistic struct {
	Rule          string       `json:"rule"`
	UploadTotal   atomic.Int64 `json:"upload"`
	DownloadTotal atomic.Int64 `json:"download"`
}

func (rs *RuleStatistic) AddDownload(size int64) {
	rs.DownloadTotal.Add(size)
}

func (rs *RuleStatistic) AddUpload(size int64) {
	rs.UploadTotal.Add(size)
}
