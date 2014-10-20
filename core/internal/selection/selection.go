package selection

import (
	"errors"
	"fmt"
	"github.com/sclevine/agouti/core/internal/types"
	"strings"
)

type Selection struct {
	Client    client
	selectors []types.Selector
	acceptAll bool
}

type client interface {
	DoubleClick() error
	MoveTo(element types.Element, point types.Point) error
	retriever
}

func (s *Selection) At(index int) types.Selection {
	last := len(s.selectors) - 1

	if last < 0 {
		return &Selection{s.Client, newSelectors(s.selectors), s.acceptAll}
	}

	newSelector := s.selectors[last]
	newSelector.Index = index
	newSelector.Indexed = true
	return &Selection{s.Client, newSelectors(s.selectors[:last], newSelector), s.acceptAll}
}

func (s *Selection) Find(selector string) types.Selection {
	last := len(s.selectors) - 1
	if last >= 0 && s.selectors[last].Using == "css selector" && !s.selectors[last].Indexed {
		return s.mergedSelection(selector)
	}

	return s.subSelection("css selector", selector)
}

func (s *Selection) FindXPath(selector string) types.Selection {
	return s.subSelection("xpath", selector)
}

func (s *Selection) FindLink(text string) types.Selection {
	return s.subSelection("link text", text)
}

func (s *Selection) FindByLabel(text string) types.Selection {
	selector := fmt.Sprintf(`//input[@id=(//label[normalize-space(text())="%s"]/@for)] | //label[normalize-space(text())="%s"]/input`, text, text)
	return s.FindXPath(selector)
}

func (s *Selection) All() types.Selection {
	return &Selection{s.Client, newSelectors(s.selectors), true}
}

func (s *Selection) subSelection(using, value string) *Selection {
	newSelector := types.Selector{Using: using, Value: value}
	return &Selection{s.Client, newSelectors(s.selectors, newSelector), s.acceptAll}
}

func (s *Selection) mergedSelection(value string) *Selection {
	last := len(s.selectors) - 1
	newSelectorValue := s.selectors[last].Value + " " + value
	newSelector := types.Selector{Using: "css selector", Value: newSelectorValue}
	return &Selection{s.Client, newSelectors(s.selectors[:last], newSelector), s.acceptAll}
}

func newSelectors(selectors []types.Selector, newSelectors ...types.Selector) []types.Selector {
	selectorsCopy := append([]types.Selector(nil), selectors...)
	return append(selectorsCopy, newSelectors...)
}

func (s *Selection) String() string {
	var tags []string

	for _, selector := range s.selectors {
		tags = append(tags, selector.String())
	}

	selection := strings.Join(tags, " | ")
	if s.acceptAll {
		return selection + " - All"
	} else {
		return selection
	}
}

func (s *Selection) Count() (int, error) {
	elements, err := s.getElements()
	if err != nil {
		return 0, fmt.Errorf("failed to select '%s': %s", s, err)
	}

	return len(elements), nil
}

func (s *Selection) EqualsElement(comparable interface{}) (bool, error) {
	element, err := s.getSelectedElement()
	if err != nil {
		return false, fmt.Errorf("failed to select '%s': %s", s, err)
	}

	selection, ok := comparable.(*Selection)
	if !ok {
		return false, errors.New("provided object is not a selection")
	}

	otherElement, err := selection.getSelectedElement()
	if err != nil {
		return false, fmt.Errorf("failed to select '%s': %s", comparable, err)
	}

	equal, err := element.IsEqualTo(otherElement)
	if err != nil {
		return false, fmt.Errorf("failed to compare '%s' to '%s': %s", s, comparable, err)
	}

	return equal, nil
}
