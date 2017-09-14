package quiz

type Section struct {
	HasIdAndTitle
	Questions   []*QuestionAndAnswer `json:"questions,omitempty" xml:"question"`
	SubSections []*SubSection        `json:"subSections,omitempty" xml:"subsection"`

	DefaultChoices []*Text           `json:"defaultChoices,omitempty" xml:"default_choices"`
	AnswersAsChoices bool            `json:"answersAsChoices" xml:"answers_as_choices,attr"`

	// These do not appear in the JSON.
	subSectionsMap map[string]*SubSection `json:"-" xml:"-"`
	CountQuestions int                    `json:"-" xml:"-"`

	// An array of all questions in all sub-sections and in this section directly.
	QuestionsArray []*QuestionAndAnswer `json:"-" xml:"-"`
}


func (self *Section) GetSubSection(subSectionId string) *SubSection {
	if self.subSectionsMap == nil {
		return nil
	}

	s, ok := self.subSectionsMap[subSectionId]
	if (!ok) {
		return nil
	}

	if (s == nil) {
		return nil
	}

	return s
}
