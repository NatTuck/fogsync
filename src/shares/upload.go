package shares

func (ss *Share) upload() {
	ss.Uploads<- true
}

func (ss *Share) uploadLoop() {
	for {
		select {
		case _ = <-ss.Uploads:
			settings := config.GetSettings()
			if !settings.Ready() {
				continue
			}

			ss.reallyUpload()
		}
	}
}

func (ss *Share) reallyUpload() {
	
}
