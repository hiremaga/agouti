package selection_test

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti/core/internal/mocks"
	. "github.com/sclevine/agouti/core/internal/selection"
	"github.com/sclevine/agouti/core/internal/types"
)

var _ = Describe("Selection", func() {
	var (
		selection types.Selection
		client    *mocks.Client
		element   *mocks.Element
	)

	BeforeEach(func() {
		client = &mocks.Client{}
		element = &mocks.Element{}
		selection = &Selection{Client: client}
		selection = selection.Find("#selector")
	})

	ItShouldEnsureASingleElement := func(matcher func() error) {
		Context("ensures a single element is returned", func() {
			It("should return an error with the number of elements", func() {
				client.GetElementsCall.ReturnElements = []types.Element{element, element}
				Expect(matcher()).To(MatchError("failed to retrieve element with 'CSS: #selector': multiple elements (2) were selected"))
			})
		})
	}

	Describe("#Find", func() {
		Context("when there is no selection", func() {
			It("should add a new css selector to the selection", func() {
				selection := &Selection{Client: client}
				Expect(selection.Find("#selector").String()).To(Equal("CSS: #selector"))
			})
		})

		Context("when the selection ends with an xpath selector", func() {
			It("should add a new css selector to the selection", func() {
				xpath := selection.FindXPath("//subselector")
				Expect(xpath.Find("#subselector").String()).To(Equal("CSS: #selector | XPath: //subselector | CSS: #subselector"))
			})
		})

		Context("when the selection ends with an unindexed CSS selector", func() {
			It("should modifie the terminal css selector to include the new selector", func() {
				Expect(selection.Find("#subselector").String()).To(Equal("CSS: #selector #subselector"))
			})
		})

		Context("when the selection ends with an indexed CSS selector", func() {
			It("should add a new css selector to the selection", func() {
				Expect(selection.At(0).Find("#subselector").String()).To(Equal("CSS: #selector [0] | CSS: #subselector"))
			})
		})

		Context("when two CSS selections are created from the same XPath parent", func() {
			It("should not overwrite the first created child", func() {
				selection := &Selection{Client: client}
				parent := selection.FindXPath("//one").FindXPath("//two").FindXPath("//parent")
				firstChild := parent.Find("#firstChild")
				parent.Find("#secondChild")
				Expect(firstChild.String()).To(Equal("XPath: //one | XPath: //two | XPath: //parent | CSS: #firstChild"))
			})
		})
	})

	Describe("#FindXPath", func() {
		It("should add a new XPath selector to the selection", func() {
			Expect(selection.FindXPath("//subselector").String()).To(Equal("CSS: #selector | XPath: //subselector"))
		})
	})

	Describe("#FindLink", func() {
		It("should add a new 'link text' selector to the selection", func() {
			Expect(selection.FindLink("some text").String()).To(Equal(`CSS: #selector | Link: "some text"`))
		})
	})

	Describe("#FindByLabel", func() {
		It("should add an XPath selector for finding by label", func() {
			Expect(selection.FindByLabel("label name").String()).To(Equal(`CSS: #selector | XPath: //input[@id=(//label[normalize-space(text())="label name"]/@for)] | //label[normalize-space(text())="label name"]/input`))
		})
	})

	Describe("#All", func() {
		It("should return a MultiSelection created from the Selection", func() {
			Expect(selection.All().String()).To(Equal(`CSS: #selector - All`))
		})
	})

	Describe("#String", func() {
		It("should return the separated selectors", func() {
			Expect(selection.FindXPath("//subselector").String()).To(Equal("CSS: #selector | XPath: //subselector"))
		})

		Context("when indexed via At(index)", func() {
			It("should append [index] to the indexed selectors", func() {
				Expect(selection.At(2).FindXPath("//subselector").At(1).String()).To(Equal("CSS: #selector [2] | XPath: //subselector [1]"))
			})
		})
	})

	Describe("#Count", func() {
		BeforeEach(func() {
			client.GetElementsCall.ReturnElements = []types.Element{element, element}
		})

		It("should request elements from the client using the provided selector", func() {
			selection.Count()
			Expect(client.GetElementsCall.Selector).To(Equal(types.Selector{Using: "css selector", Value: "#selector"}))
		})

		Context("when the client succeeds in retrieving the elements", func() {
			It("should return the text", func() {
				count, _ := selection.Count()
				Expect(count).To(Equal(2))
			})

			It("should not return an error", func() {
				_, err := selection.Count()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the the client fails to retrieve the elements", func() {
			BeforeEach(func() {
				client.GetElementsCall.Err = errors.New("some error")
			})

			It("should return an error", func() {
				_, err := selection.Count()
				Expect(err).To(MatchError("failed to retrieve elements for 'CSS: #selector': some error"))
			})
		})
	})

	Describe("#EqualsElement", func() {
		var (
			otherClient    *mocks.Client
			otherSelection types.Selection
			otherElement   *mocks.Element
		)

		BeforeEach(func() {
			client.GetElementsCall.ReturnElements = []types.Element{element}
			otherClient = &mocks.Client{}
			otherSelection = &Selection{Client: otherClient}
			otherSelection = otherSelection.Find("#other_selector")
			otherElement = &mocks.Element{}
			otherClient.GetElementsCall.ReturnElements = []types.Element{otherElement}
		})

		ItShouldEnsureASingleElement(func() error {
			_, err := selection.EqualsElement(otherSelection)
			return err
		})

		It("should ensure that the other selection is a single element", func() {
			otherClient.GetElementsCall.ReturnElements = []types.Element{element, element}
			_, err := selection.EqualsElement(otherSelection)
			Expect(err).To(MatchError("failed to retrieve element with 'CSS: #other_selector': multiple elements (2) were selected"))
		})

		It("should compare the selection elements for equality", func() {
			selection.EqualsElement(otherSelection)
			Expect(element.IsEqualToCall.Element).To(Equal(otherElement))
		})

		Context("if the provided element is not a *Selection", func() {
			It("should return an error", func() {
				_, err := selection.EqualsElement("not a selection")
				Expect(err).To(MatchError("provided object is not a selection"))
			})
		})

		Context("if the client fails to compare the elements", func() {
			It("should return an error", func() {
				element.IsEqualToCall.Err = errors.New("some error")
				_, err := selection.EqualsElement(otherSelection)
				Expect(err).To(MatchError("failed to compare 'CSS: #selector' to 'CSS: #other_selector': some error"))
			})
		})

		Context("if the client succeeds in comparing the elements", func() {
			It("should return true if they are equal", func() {
				element.IsEqualToCall.ReturnEquals = true
				equal, _ := selection.EqualsElement(otherSelection)
				Expect(equal).To(BeTrue())
			})

			It("should return false if they are not equal", func() {
				element.IsEqualToCall.ReturnEquals = false
				equal, _ := selection.EqualsElement(otherSelection)
				Expect(equal).To(BeFalse())
			})

			It("should not return an error", func() {
				_, err := selection.EqualsElement(otherSelection)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
