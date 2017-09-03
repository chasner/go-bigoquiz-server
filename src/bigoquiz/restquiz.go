package bigoquiz

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"quiz"
	"strconv"
	"strings"
)

func filesWithExtension(dirPath string, ext string) ([]string, error) {
	result := make([]string, 0)

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		fmt.Println(err)
		return result, err
	}

	dotSuffix := "." + ext
	suffixLen := len(dotSuffix)
	for _, f := range files {

		name := f.Name()
		if strings.HasSuffix(name, dotSuffix) {
			prefix := name[0 : len(name)-suffixLen]
			result = append(result, prefix)
		}
	}

	return result, nil
}

func loadQuiz(id string) (*quiz.Quiz, error) {
	absFilePath, err := filepath.Abs("quizzes/" + id + ".xml")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return quiz.LoadQuiz(absFilePath, id)
}

func loadQuizzes() (map[string]*quiz.Quiz, error) {
	quizzes := make(map[string]*quiz.Quiz, 0)

	absFilePath, err := filepath.Abs("quizzes")
	if err != nil {
		fmt.Println(err)
		return quizzes, err
	}

	quizNames, err := filesWithExtension(absFilePath, "xml")
	if err != nil {
		fmt.Println(err)
		return quizzes, err
	}

	for _, name := range quizNames {
		q, err := loadQuiz(name)
		if err != nil {
			fmt.Println(err)
			return quizzes, err
		}

		quizzes[q.Id] = q
	}

	return quizzes, nil
}

// TODO: Is there instead some way to output just the top-level of the JSON,
// and only some of the fields?
func buildQuizzesSimple(quizzes map[string]*quiz.Quiz) []*quiz.Quiz {
	// Create a slice with the same capacity.
	result := make([]*quiz.Quiz, 0, len(quizzes))

	for _, q := range quizzes {
		var simple quiz.Quiz
		q.CopyHasIdAndTitle(&simple.HasIdAndTitle)
		simple.IsPrivate = q.IsPrivate

		result = append(result, &simple)
	}

	return result
}

func buildQuizzesFull(quizzes map[string]*quiz.Quiz) []*quiz.Quiz {
	// Create a slice with the same capacity.
	result := make([]*quiz.Quiz, 0, len(quizzes))

	for _, q := range quizzes {
		result = append(result, q)
	}

	return result
}

func restHandleQuizAll(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	listOnly := false
	queryValues := r.URL.Query()
	if queryValues != nil {
		listOnlyStr := queryValues.Get("list_only")
		listOnly, _ = strconv.ParseBool(listOnlyStr)
	}

	// TODO: Cache this.
	quizzes, err := loadQuizzes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var quizArray []*quiz.Quiz = nil
	if listOnly {
		// TODO: Cache this.
		quizArray = buildQuizzesSimple(quizzes)
	} else {
		// TODO: Cache this.
		quizArray = buildQuizzesFull(quizzes)
	}

	w.Header().Set("Content-Type", "application/json") // normal header
	w.WriteHeader(http.StatusOK)

	jsonStr, err := json.Marshal(quizArray)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonStr)
}

func getQuiz(quizId string) *quiz.Quiz {
	// TODO: Cache this.
	quizzes, err := loadQuizzes()
	if err != nil {
		return nil
	}

	return quizzes[quizId]
}

func restHandleQuizById(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	quizId := ps.ByName("quizId")
	if quizId == "" {
		// This makes no sense. restHandleQuizAll() should have been called.
		http.Error(w, "Empty quiz ID", http.StatusInternalServerError)
		return
	}

	q := getQuiz(quizId)
	if q == nil {
		http.Error(w, "quiz not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json") // normal header
	w.WriteHeader(http.StatusOK)

	jsonStr, err := json.Marshal(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonStr)
}

func restHandleQuizSectionsByQuizId(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	listOnly := false
	queryValues := r.URL.Query()
	if queryValues != nil {
		listOnlyStr := queryValues.Get("list_only")
		listOnly, _ = strconv.ParseBool(listOnlyStr)
	}

	quizId := ps.ByName("quizId")
	if quizId == "" {
		// This makes no sense. restHandleQuizAll() should have been called.
		http.Error(w, "Empty quiz ID", http.StatusInternalServerError)
		return
	}

	q := getQuiz(quizId)
	if q == nil {
		http.Error(w, "quiz not found", http.StatusInternalServerError)
		return
	}

	sections := q.Sections
	if listOnly {
		simpleSections := make([]*quiz.Section, 0, len(sections))
		for _, s := range sections {
			var simple quiz.Section
			s.CopyHasIdAndTitle(&simple.HasIdAndTitle)
			simpleSections = append(simpleSections, &simple)
		}

		sections = simpleSections
	}

	jsonStr, err := json.Marshal(sections)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonStr)
}

func restHandleQuizQuestionById(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	quizId := ps.ByName("quizId")
	if quizId == "" {
		// This makes no sense. restHandleQuizAll() should have been called.
		http.Error(w, "Empty quiz ID", http.StatusInternalServerError)
		return
	}

	questionId := ps.ByName("questionId")
	if questionId == "" {
		// This makes no sense.
		http.Error(w, "Empty question ID", http.StatusInternalServerError)
	}

	q := getQuiz(quizId)
	if q == nil {
		http.Error(w, "quiz not found", http.StatusNotFound)
		return
	}

	qa := q.GetQuestionAndAnswer(questionId)
	if qa == nil {
		http.Error(w, "question not found", http.StatusInternalServerError)
	}

	jsonStr, err := json.Marshal(qa.Question)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonStr)
}
