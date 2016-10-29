package handler

import (
	"testing"
)


func TestNewMtgSearch(t *testing.T) {
	msr := NewMtgSearch()

	if msr.db == nil {
		t.Fatalf("Loading database failed!")
	}
}

func TestMatch(t *testing.T) {
	msr := NewMtgSearch()
	should := "[[Bloodsoaked Champion]]"
	shouldNot := "#[[mtgstatsrequest]]"
	multiShould := `And then I said [[Bloodsoaked Champion]] was the best
Mardu card. Then he said [[Progenitus]] was. I'm like wtf dude [[Emrakul the Aeons Torn]]
is even better!`
	someShouldSomeNot := `What the fuck is this about?! #[[stats]] [[Emrakul]] [[bitch]] #[[more stats]]`
	if !msr.Match(should) {
		t.Fatalf("msr.Match failed on %s\n", should)
	}
	t.Log(msr.matches)

	if msr.Match(shouldNot) {
		t.Fatalf("msr.Match passed on %s\n", shouldNot)
	}
	t.Log(msr.matches)

	if !msr.Match(multiShould) || len(msr.matches) != 3 {
		t.Fatalf("msr.Match failed to match all results in %s: %v", multiShould, msr.matches)
	}
	t.Log(msr.matches)

	if !msr.Match(someShouldSomeNot) || len(msr.matches) < 2 {
		t.Fatalf("msr.Match failed to match all results in %s: %v", someShouldSomeNot, msr.matches)
	} else if len(msr.matches) > 2 {
		t.Fatalf("msr.Match matched too many results in %s : %v", someShouldSomeNot, msr.matches)
	}

	t.Log(msr.matches)
}

func TestRespond(t *testing.T) {
	msr := NewMtgSearch()
	should := "[[Bloodsoaked Champion]]"
	shouldNot := "[[no such card]]"
	multiShould := `And then I said [[Bloodsoaked Champion]] was the best
Mardu card. Then he said [[Progenitus]] was. I'm like wtf dude [[Emrakul the Aeons Torn]]
is even better!`
	someShouldSomeNot := `What the fuck is this about?! [[Muscle Burst]] [[Emrakul]] [[bitch]] [[more stats]]`

	msr.Match(should)
	s, err := msr.Respond()
	t.Log(s)

	msr.Match(shouldNot)
	s, err = msr.Respond()
	t.Log(s)

	msr.Match(multiShould)
	s, err = msr.Respond()
	t.Log(s)

	msr.Match(someShouldSomeNot)
	s, err = msr.Respond()
	t.Log(s)

	msr.Match("[[Glorious Anthem|7ED]]")
	s, err = msr.Respond()
	t.Log(s)

	if err != nil {
		t.Fatalf("whoops! %s")
	}
}
