package downloader

type Downloader interface {
	DownloadExec(location string, args string) string
}

func NewHttpDownloader(agent string) *HttpDownload {
	h := new(HttpDownload)

	h.agent = agent

	return h
}
