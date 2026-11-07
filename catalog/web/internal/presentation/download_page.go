package presentation

func NativeDownloadURL(page *ExportPage) string {
	if page == nil {
		return "/download"
	}
	return "/download?" + InitialExportDialog(*page).Query.Encode()
}
