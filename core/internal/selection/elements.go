package selection

import (
	"errors"
	"fmt"

	"github.com/sclevine/agouti/core/internal/types"
)

type retriever interface {
	GetElements(selector types.Selector) ([]types.Element, error)
}

func getElements(retriever retriever, selector types.Selector) ([]types.Element, error) {
	elements, err := retriever.GetElements(selector)
	if err != nil {
		return nil, err
	}

	if selector.Indexed {
		if selector.Index >= len(elements) {
			return nil, fmt.Errorf("element index out of range (>%d)", len(elements)-1)
		}

		elements = []types.Element{elements[selector.Index]}
	}

	return elements, nil
}

func (s *Selection) getElements() ([]types.Element, error) {
	if len(s.selectors) == 0 {
		return nil, errors.New("empty selection")
	}

	lastElements, err := getElements(s.Client, s.selectors[0])
	if err != nil {
		return nil, err
	}

	for _, selector := range s.selectors[1:] {
		elements := []types.Element{}
		for _, element := range lastElements {
			subElements, err := getElements(element, selector)
			if err != nil {
				return nil, err
			}

			elements = append(elements, subElements...)
		}
		lastElements = elements
	}
	return lastElements, nil
}

func (s *Selection) getSelectedElements() ([]types.Element, error) {
	elements, err := s.getElements()
	if err != nil {
		return nil, err
	}

	if len(elements) == 0 {
		return nil, fmt.Errorf("no elements found")
	}

	// TODO: never tested
	if len(elements) > 1 && !s.acceptAll  {
		return nil, fmt.Errorf("method requires All() for multiple elements (%d) ", len(elements))
	}

	return elements, nil
}

// TODO: never tested
func (s *Selection) getSelectedElement() (types.Element, error) {
	if s.acceptAll {
		return nil, fmt.Errorf("method does not support All()")
	}

	elements, err := s.getElements()
	if err != nil {
		return nil, err
	}

	if len(elements) == 0 {
		return nil, fmt.Errorf("no elements found")
	}

	if len(elements) > 1 {
		return nil, fmt.Errorf("method does not support multiple elements (%d) ", len(elements))
	}

	return elements[0], nil
}