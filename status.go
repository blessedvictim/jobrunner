package jobrunner

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/robfig/cron/v3"
)

type StatusData struct {
	Id        cron.EntryID
	JobRunner *Job
	Next      time.Time
	Prev      time.Time
}

// Return detailed list of currently running recurring jobs
// to remove an entry, first retrieve the ID of entry
func Entries() []cron.Entry {
	return MainCron.Entries()
}

func StatusPage() []StatusData {

	ents := MainCron.Entries()

	Statuses := make([]StatusData, len(ents))
	for k, v := range ents {
		Statuses[k].Id = v.ID
		Statuses[k].JobRunner = AddJob(v.Job)
		Statuses[k].Next = v.Next
		Statuses[k].Prev = v.Prev

	}

	// t := template.New("status_page")

	// var data bytes.Buffer
	// t, _ = t.ParseFiles("views/Status.html")

	// t.ExecuteTemplate(&data, "status_page", Statuses())
	return Statuses
}

func StatusJson() map[string]interface{} {

	return map[string]interface{}{
		"jobrunner": StatusPage(),
	}

}

type JobToRunNow struct {
	JobID int `json:"job_id"`
}

type JsonError struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func HandleRunJob(req *http.Request, w http.ResponseWriter) {
	var args JobToRunNow
	err := json.NewDecoder(req.Body).Decode(&args)
	if err != nil {
		b, _ := json.Marshal(JsonError{
			Error: "decode failed",
			Code:  "1",
		})
		_, _ = w.Write(b)
	}

	StartJobNow(cron.EntryID(args.JobID))
}

func StartJobNow(id cron.EntryID) {
	j := MainCron.Entry(id).Job.(*Job)
	go j.Run()
}

func AddJob(job cron.Job) *Job {
	return job.(*Job)
}
