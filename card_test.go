// Copyright © 2016 Aaron Longwell
//
// Use of this source code is governed by an MIT license.
// Details in the LICENSE file.

package trello

import (
	"net/http"
	"testing"
	"time"
)

func TestCardCreatedAt(t *testing.T) {
	c := Card{}
	c.ID = "4d5ea62fd76aa1136000000c"
	ts := c.CreatedAt()
	if ts.IsZero() {
		t.Error("Time shouldn't be zero.")
	}
	if ts.Unix() != 1298048559 {
		t.Errorf("Incorrect CreatedAt() time: '%s'.", ts.Format(time.RFC3339))
	}
}

func TestGetCardsOnBoard(t *testing.T) {
	board := testBoard(t)

	server := NewMockResponder(t)
	defer server.Close()

	board.client.BaseURL = server.URL()
	cards, err := board.GetCards(Defaults())
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 5 {
		t.Errorf("Expected 5 cards, got %d", len(cards))
	}
}

func TestGetCardsInList(t *testing.T) {
	list := testList(t)

	server := NewMockResponder(t, "cards", "list-cards-api-example.json")
	defer server.Close()
	list.client.BaseURL = server.URL()

	cards, err := list.GetCards(Defaults())
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 1 {
		t.Errorf("Expected 1 cards, got %d", len(cards))
	}
}

func TestCardsCustomFields(t *testing.T) {
	list := testList(t)

	server := NewMockResponder(t, "cards", "list-cards-api-example.json")
	defer server.Close()
	list.client.BaseURL = server.URL()

	cards, err := list.GetCards(Defaults())
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 1 {
		t.Errorf("Expected 1 cards, got %d", len(cards))
	}

	if len(cards[0].CustomFieldItems) != 2 {
		t.Errorf("Expected 2 custom field items on card %s, got %d", cards[0].ID, len(cards[0].CustomFieldItems))
	}

	customFields := testBoardCustomFields(t)
	fields := cards[0].CustomFields(customFields)

	if len(fields) != 2 {
		t.Errorf("Expected 2 map items on parsed custom fields")
	}

	vf1, ok := fields["Field1"]
	if !ok || vf1 != "F1 1st opt" {
		t.Errorf("Expected Field1 to be 'F1 1st opt' but it was %v", vf1)
	}

	vf2, ok := fields["Field2"]
	if !ok || vf2 != "F2 2nd opt" {
		t.Errorf("Expected Field1 to be 'F2 2nd opt' but it was %v", vf2)
	}
}

func TestBoardContainsCopyOfCard(t *testing.T) {
	board := testBoard(t)

	server := NewMockResponder(t, "actions", "board-actions-copyCard.json")
	defer server.Close()

	board.client.BaseURL = server.URL()
	firstResult, err := board.ContainsCopyOfCard("57f50c552b96e3fffe588aad", Defaults())
	if err != nil {
		t.Error(err)
	}
	if firstResult {
		t.Errorf("Incorrect Copy test: Card 57f50c552b96e3fffe588aad was never copied.")
	}

	secondResult, err := board.ContainsCopyOfCard("57914873fd2de1a10f3cb422", Defaults())
	if err != nil {
		t.Error(err)
	}
	if !secondResult {
		t.Errorf("ContainsCopyOfCard(57f50c552b96e3fffe588aad) should have been true.")
	}
}

func TestCreateCard(t *testing.T) {
	c := testClient()
	server := NewMockResponder(t, "cards", "card-create.json")
	defer server.Close()
	server.AssertRequest(func(t *testing.T, r *http.Request) {
		due := r.URL.Query().Get("due")
		if _, err := time.Parse(time.RFC3339, due); err != nil {
			t.Errorf("Expected due to be in RFC3339 format, but value was '%v'", due)
		}

		start := r.URL.Query().Get("start")
		if _, err := time.Parse(time.RFC3339, start); err != nil {
			t.Errorf("Expected start to be in RFC3339 format, but value was '%v'", start)
		}
	})

	c.BaseURL = server.URL()
	dueDate := time.Now().AddDate(0, 0, 3)
	startDate := time.Now().AddDate(0, 0, 2)

	card := Card{
		Name:     "Test Card Create",
		Desc:     "What its about",
		Due:      &dueDate,
		Start:    &startDate,
		IDList:   "57f03a06b5ff33a63c8be316",
		IDLabels: []string{"label1", "label2"},
	}

	err := c.CreateCard(&card, Arguments{"pos": "top"})
	if err != nil {
		t.Error(err)
	}

	if card.Pos != 8192 {
		t.Errorf("Expected card to pick up a new Pos value. Instead got %.2f.", card.Pos)
	}

	if card.DateLastActivity == nil {
		t.Error("Expected card to pick up a last activity date. Was nil.")
	}

	if card.ID != "57f5183c691585658d408681" {
		t.Errorf("Expected card to pick up an ID. Instead got '%s'.", card.ID)
	}

	if len(card.Labels) < 2 {
		t.Errorf("Expected card to be assigned two labels. Instead got '%v'.", card.Labels)
	}
}

func TestSetCustomField(t *testing.T) {
	c := testClient()
	server := mockResponder{t: t}

	defer server.Close()

	c.BaseURL = server.URL()
	cfItem := CustomFieldItem{
		Value:         NewCustomFieldValue("probando"),
		IDModel:       "57f5183c691585658d408681",
		IDCustomField: "57f5183c691585658d408681",
	}

	err := c.SetCustomFieldByItem(cfItem)
	if err != nil {
		t.Error(err)
	}

	err = c.SetCustomField(cfItem.ID, cfItem.IDCustomField, "otra prueba")
	if err != nil {
		t.Error(err)
	}
}

func TestAddCardToList(t *testing.T) {
	l := testList(t)

	server := NewMockResponder(t, "cards", "card-posted-to-bottom-of-list.json")
	server.AssertRequest(func(t *testing.T, r *http.Request) {
		due := r.URL.Query().Get("due")
		if _, err := time.Parse(time.RFC3339, due); err != nil {
			t.Errorf("Expected due to be in RFC3339 format, but value was '%v'", due)
		}

		start := r.URL.Query().Get("start")
		if _, err := time.Parse(time.RFC3339, start); err != nil {
			t.Errorf("Expected start to be in RFC3339 format, but value was '%v'", start)
		}
	})
	defer server.Close()
	l.client.BaseURL = server.URL()
	dueDate := time.Now().AddDate(0, 0, 2)
	startDate := time.Now().AddDate(0, 0, 1)

	card := Card{
		Name:     "Test Card POSTed to List",
		Desc:     "This is its description.",
		Due:      &dueDate,
		Start:    &startDate,
		IDLabels: []string{"label1", "label2"},
	}

	err := l.AddCard(&card, Arguments{"pos": "bottom"})
	if err != nil {
		t.Error(err)
	}

	if card.Pos != 32768 {
		t.Errorf("Expected card to pick up a new Pos value. Instead got %.2f.", card.Pos)
	}

	if card.DateLastActivity == nil {
		t.Error("Expected card to pick up a last activity date. Was nil.")
	}

	if card.ID != "57f5118667db8839dab68698" {
		t.Errorf("Expected card to pick up an ID. Instead got '%s'.", card.ID)
	}

	if len(card.Labels) < 2 {
		t.Errorf("Expected card to be assigned two labels. Instead got '%v'.", card.Labels)
	}
}

func TestArchiveUnarchive(t *testing.T) {
	c := testCard(t)

	server := NewMockResponder(t, "cards", "card-archived.json")
	c.client.BaseURL = server.URL()
	c.Archive()
	if c.Closed == false {
		t.Errorf("Card should have been archived.")
	}
	server.Close()

	server = NewMockResponder(t, "cards", "card-unarchived.json")
	c.client.BaseURL = server.URL()
	c.Unarchive()
	if c.Closed == true {
		t.Errorf("Card should have been unarchived.")
	}
	server.Close()
}

func TestCopyCardToList(t *testing.T) {
	c := testCard(t)

	server := NewMockResponder(t, "cards", "card-copied.json")
	defer server.Close()
	c.client.BaseURL = server.URL()

	newCard, err := c.CopyToList("57f03a022cd45c863ca581f1", Defaults())
	if err != nil {
		t.Error(err)
	}

	if newCard.ID == c.ID {
		t.Errorf("New card should have a new ID: '%s'.", newCard.ID)
	}

	if newCard.Pos != 16384 {
		t.Errorf("Expected new card to have correct Pos value. Got %.2f", newCard.Pos)
	}
}

func TestGetParentCard(t *testing.T) {
	c := testCard(t)

	server := NewMockResponder(t)
	defer server.Close()
	c.client.BaseURL = server.URL()

	parent, err := c.GetParentCard(Defaults())
	if err != nil {
		t.Error(err)
	}
	if parent == nil {
		t.Errorf("Problem")
	}
}

func TestGetAncestorCards(t *testing.T) {
	c := testCard(t)

	server := mockDynamicPathResponse()
	defer server.Close()
	c.client.BaseURL = server.URL

	ancestors, err := c.GetAncestorCards(Defaults())
	if err != nil {
		t.Error(err)
	}
	if len(ancestors) != 1 {
		t.Errorf("Expected 1 ancestor, got %d", len(ancestors))
	}
}

func TestAddMemberIdToCard(t *testing.T) {
	c := testCard(t)
	server := NewMockResponder(t, "cards", "card-add-member-response.json")
	defer server.Close()

	c.client.BaseURL = server.URL()
	member, err := c.AddMemberID("testmemberid")
	if err != nil {
		t.Error(err)
	}
	if member[0].ID != "testmemberid" {
		t.Errorf("Expected id testmemberid, got %v", member[0].ID)
	}
	if member[0].Username != "testmemberusername" {
		t.Errorf("Expected username testmemberusername, got %v", member[0].Username)
	}
}

func TestAddURLAttachmentToCard(t *testing.T) {
	c := testCard(t)
	server := NewMockResponder(t, "cards", "url-attachments.json")
	defer server.Close()

	c.client.BaseURL = server.URL()
	attachment := Attachment{
		Name: "Test",
		URL:  "https://github.com/test",
	}
	err := c.AddURLAttachment(&attachment)
	if err != nil {
		t.Error(err)
	}
	if attachment.ID != "5bbce18fa4a337483b145a57" {
		t.Errorf("Expected attachment to pick up an ID, got %v instead", attachment.ID)
	}
}

func TestCardSetClient(t *testing.T) {
	card := Card{}
	client := testClient()
	card.SetClient(client)
	if card.client == nil {
		t.Error("Expected non-nil card.client")
	}
}

// Utility function to get a simple response from Client.GetCard()
func testCard(t *testing.T) *Card {
	c := testClient()
	server := NewMockResponder(t, "cards", "card-api-example.json")
	defer server.Close()

	c.BaseURL = server.URL()
	card, err := c.GetCard("4eea503", Defaults())
	if err != nil {
		t.Fatal(err)
	}
	return card
}
