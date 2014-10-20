package selection_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti/core/internal/selection"
	"github.com/sclevine/agouti/core/internal/mocks"
	"github.com/sclevine/agouti/core/internal/types"
)

var _ = Describe("Elements", func() {

	// TODO: test mergedSelection copying and At copying in two places
	// TODO: figure out how to separate single/multiple elements

	Describe("#At + most methods: retrieving elements", func() {
		var (
			parentOne *mocks.Element
			parentTwo *mocks.Element
			count     int
		)

		BeforeEach(func() {
			parentOne = &mocks.Element{}
			parentTwo = &mocks.Element{}
			parentOne.GetElementsCall.ReturnElements = []types.Element{&mocks.Element{}, &mocks.Element{}}
			parentTwo.GetElementsCall.ReturnElements = []types.Element{&mocks.Element{}, &mocks.Element{}}
			client.GetElementsCall.ReturnElements = []types.Element{parentOne, parentTwo}
		})

		Context("when successful without indices", func() {
			BeforeEach(func() {
				selection = selection.FindXPath("children")
				count, _ = selection.Count()
			})

			It("should retrieve the parent elements using the client", func() {
				Expect(client.GetElementsCall.Selector).To(Equal(types.Selector{Using: "css selector", Value: "#selector"}))
			})

			It("should retrieve the child elements of the parent selector", func() {
				Expect(parentOne.GetElementsCall.Selector).To(Equal(types.Selector{Using: "xpath", Value: "children"}))
				Expect(parentTwo.GetElementsCall.Selector).To(Equal(types.Selector{Using: "xpath", Value: "children"}))
			})

			It("should return all child elements of the terminal selector", func() {
				Expect(count).To(Equal(4))
			})
		})

		Context("when successful with indices", func() {
			BeforeEach(func() {
				selection.At(1).FindXPath("children").At(1).Click()
			})

			It("should retrieve the parent elements using the client", func() {
				Expect(client.GetElementsCall.Selector).To(Equal(types.Selector{Using: "css selector", Value: "#selector", Index: 1, Indexed: true}))
			})

			It("should retrieve the child elements of the parent selector", func() {
				Expect(parentOne.GetElementsCall.Selector.Using).To(BeEmpty())
				Expect(parentTwo.GetElementsCall.Selector).To(Equal(types.Selector{Using: "xpath", Value: "children", Index: 1, Indexed: true}))
			})

			It("should return all child elements of the terminal selector", func() {
				clickedElement := parentTwo.GetElementsCall.ReturnElements[1].(*mocks.Element)
				Expect(clickedElement.ClickCall.Called).To(BeTrue())
			})
		})

		Context("when there is no selection", func() {
			BeforeEach(func() {
				selection = &Selection{Client: client}
			})

			It("should return an error", func() {
				_, err := selection.Count()
				Expect(err).To(MatchError("failed to retrieve elements for '': empty selection"))
			})
		})

		Context("when retrieving the parent elements fails", func() {
			BeforeEach(func() {
				selection = selection.FindXPath("children")
				client.GetElementsCall.Err = errors.New("some error")
			})

			It("should return the error", func() {
				_, err := selection.Count()
				Expect(err).To(MatchError("failed to retrieve elements for 'CSS: #selector | XPath: children': some error"))
			})
		})

		Context("when retrieving any of the child elements fails", func() {
			BeforeEach(func() {
				selection = selection.FindXPath("children")
				parentTwo.GetElementsCall.Err = errors.New("some error")
			})

			It("should return the error", func() {
				_, err := selection.Count()
				Expect(err).To(MatchError("failed to retrieve elements for 'CSS: #selector | XPath: children': some error"))
			})
		})

		Context("when the first selection index is out of range", func() {
			It("should return an error with the index and total number of elements", func() {
				Expect(selection.At(2).Click()).To(MatchError("failed to retrieve element with 'CSS: #selector [2]': element index out of range (>1)"))
			})
		})

		Context("when subsequent selection indices are out of range", func() {
			It("should return an error with the index and total number of elements", func() {
				Expect(selection.At(0).Find("#selector").At(2).Click()).To(MatchError("failed to retrieve element with 'CSS: #selector [0] | CSS: #selector [2]': element index out of range (>1)"))
			})
		})
	})

	Describe("most methods: retrieving the selected element", func() {
		It("should request an element from the client using the element's selector", func() {
			selection.Click()
			Expect(client.GetElementsCall.Selector).To(Equal(types.Selector{Using: "css selector", Value: "#selector"}))
		})

		Context("when the client fails to retrieve any elements", func() {
			It("should return error from the client", func() {
				client.GetElementsCall.Err = errors.New("some error")
				Expect(selection.Click()).To(MatchError("failed to retrieve element with 'CSS: #selector': some error"))
			})
		})

		Context("when the client retrieves zero elements", func() {
			It("should fail with an error indicating there were no elements", func() {
				client.GetElementsCall.ReturnElements = []types.Element{}
				Expect(selection.Click()).To(MatchError("failed to retrieve element with 'CSS: #selector': no element found"))
			})
		})

		Context("when the client retrieves more than one element and indexing is disabled", func() {
			It("should return an error with the number of elements", func() {
				client.GetElementsCall.ReturnElements = []types.Element{element, element}
				Expect(selection.Click()).To(MatchError("failed to retrieve element with 'CSS: #selector': multiple elements (2) were selected"))
			})
		})
	})
})
