package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// GET /records/{id}
// GetRecords retrieves the latest version of record.
func (a *API) GetRecords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	idNumber, err := strconv.ParseInt(id, 10, 32)

	if err != nil || idNumber <= 0 {
		err := writeError(w, "invalid id; id must be a positive number", http.StatusBadRequest)
		logError(err)
		return
	}

	record, err := a.records.GetRecord(
		ctx,
		int(idNumber),
	)
	if err != nil {
		err := writeError(w, fmt.Sprintf("record of id %v does not exist", idNumber), http.StatusBadRequest)
		logError(err)
		return
	}

	err = writeJSON(w, record, http.StatusOK)
	logError(err)
}

// GET /records/{id}?versions={versions}
// GetRecordsV2 retrieves all the specified versions of records.
func (a *API) GetRecordsV2(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	versions := mux.Vars(r)["versions"]

	idNumber, err := strconv.ParseInt(id, 10, 32)

	if err != nil || idNumber <= 0 {
		err := writeError(w, "invalid id; id must be a positive number", http.StatusBadRequest)
		logError(err)
		return
	}

	if strings.EqualFold(versions, "all") {
		records := a.records.GetAllRecords(
			ctx,
			int(idNumber),
		)
		err = writeJSON(w, records, http.StatusOK)
		logError(err)
		return
	} else if (strings.EqualFold(versions, "latest")) {
		record, _ := a.records.GetRecord(ctx, int(idNumber))
		err = writeJSON(w, record, http.StatusOK)
		logError(err)
		return
	}

	vs := strings.Split(versions, `,`)
	var verNumbers []int64
	for _, x := range vs {
		ver, _ := strconv.ParseInt(x, 10, 64)
		verNumbers = append(verNumbers, ver)
	}

	records := a.records.GetRecordsWithVersions(
		ctx,
		int(idNumber),
	        verNumbers,
	)

	err = writeJSON(w, records, http.StatusOK)
	logError(err)
}
