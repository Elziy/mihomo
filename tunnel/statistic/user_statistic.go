package statistic

import "github.com/metacubex/mihomo/common/atomic"

type UserStatistic struct {
	User                string       `json:"user"`
	DirectUploadTotal   atomic.Int64 `json:"directUpload"`
	DirectDownloadTotal atomic.Int64 `json:"directDownload"`
	ProxyUploadTotal    atomic.Int64 `json:"proxyUpload"`
	ProxyDownloadTotal  atomic.Int64 `json:"proxyDownload"`
}

func (us *UserStatistic) AddDownload(size int64, isProxy bool) {
	if isProxy {
		us.ProxyDownloadTotal.Add(size)
	} else {
		us.DirectDownloadTotal.Add(size)
	}
}

func (us *UserStatistic) AddUpload(size int64, isProxy bool) {
	if isProxy {
		us.ProxyUploadTotal.Add(size)
	} else {
		us.DirectUploadTotal.Add(size)
	}
}
