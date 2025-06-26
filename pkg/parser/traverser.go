package parser

import (
	"encoding/xml"
	"strings"
)

// stripns removes the namespace prefix from a qualified name
func stripns(xsdType string) string {
	if xsdType == "" {
		return ""
	}
	split := strings.Split(xsdType, ":")
	if len(split) == 2 {
		return split[1]
	}
	return xsdType
}

type traverseMode int32

const (
	refResolution traverseMode = iota
	findNameByType
)

type traverser struct {
	c   *XSDSchema
	all []*XSDSchema
	tm  traverseMode
	// fields used by findNameByType mode
	typeName             string
	foundElmName         string
	conflictingTypeUsage bool
}

// NewTraverser creates a new traverser for schema resolution
func NewTraverser(c *XSDSchema, all []*XSDSchema) *traverser {
	return &traverser{
		c:   c,
		all: all,
		tm:  refResolution, // default traverse mode is refResolution
	}
}

// Traverse performs the traversal
func (t *traverser) Traverse() {
	t.tm = refResolution

	for _, ct := range t.c.ComplexTypes {
		t.traverseComplexType(ct)
	}
	for _, st := range t.c.SimpleType {
		t.traverseSimpleType(st)
	}
	for _, elm := range t.c.Elements {
		t.traverseElement(elm)
	}
}

// Given a type, check if there is an Element with that type, and return its name.
// If multiple elements with identical names of the given type are found,
// the name is returned.
// If multiple elements with different names of the given type are found,
// the original type name is returned instead.
// If no elements are found, the original type name is returned instead.
func (t *traverser) findNameByType(name string) string {
	t.initFindNameByType(name)

	// Search for elements of given type
	for _, schema := range t.all {
		for _, elm := range schema.Elements {
			t.traverseElement(elm)
		}
		for _, ct := range schema.ComplexTypes {
			t.traverseComplexType(ct)
		}
		for _, st := range schema.SimpleType {
			t.traverseSimpleType(st)
		}
	}

	// Return found element name if given type is used only once
	if len(t.foundElmName) > 0 && !t.conflictingTypeUsage {
		return t.foundElmName
	}

	// Return original type name
	// No element found or conflicting element names found
	return t.typeName
}

func (t *traverser) initFindNameByType(name string) {
	// Initialize fields for processing
	t.tm = findNameByType
	t.typeName = stripns(name)
	t.foundElmName = ""
	t.conflictingTypeUsage = false
}

func (t *traverser) traverseElements(ct []*XSDElement) {
	for _, elm := range ct {
		t.traverseElement(elm)
	}
}

func (t *traverser) traverseElement(elm *XSDElement) {
	// Check if we are in ref resolution mode
	if t.tm == refResolution && elm.Ref != "" {
		refElm := t.getGlobalElement(elm.Ref)
		if refElm != nil && refElm.Ref == "" {
			// Copy properties from referenced element
			t.traverseElement(refElm)
			elm.Name = refElm.Name
			elm.Type = refElm.Type
			elm.Nillable = refElm.Nillable
			if elm.MinOccurs == "" {
				elm.MinOccurs = refElm.MinOccurs
			}
			if elm.MaxOccurs == "" {
				elm.MaxOccurs = refElm.MaxOccurs
			}
			// Copy complex/simple type if element doesn't have its own
			if elm.ComplexType == nil {
				elm.ComplexType = refElm.ComplexType
			}
			if elm.SimpleType == nil {
				elm.SimpleType = refElm.SimpleType
			}
		}
	}

	t.findElmName(elm)

	if elm.ComplexType != nil {
		t.traverseComplexType(elm.ComplexType)
	}
	if elm.SimpleType != nil {
		t.traverseSimpleType(elm.SimpleType)
	}
}

func (t *traverser) findElmName(elm *XSDElement) {
	// Check if we are called by findNameByType
	if t.tm != findNameByType {
		return
	}

	// Conflicting type usage already detected -> no need to search any further
	if t.conflictingTypeUsage {
		return
	}

	if stripns(elm.Type) == t.typeName {
		if len(t.foundElmName) == 0 {
			// First time usage t.typeName
			t.foundElmName = elm.Name
		} else if t.foundElmName != elm.Name {
			// Duplicate use of t.typeName with different element names
			t.conflictingTypeUsage = true
		}
	}
}

func (t *traverser) traverseSimpleType(st *XSDSimpleType) {
}

func (t *traverser) traverseComplexType(ct *XSDComplexType) {
	t.traverseElements(ct.Sequence)
	t.traverseElements(ct.Choice)
	t.traverseElements(ct.SequenceChoice)
	t.traverseElements(ct.All)
	t.traverseAttributes(ct.Attributes)
	t.traverseAttributes(ct.ComplexContent.Extension.Attributes)
	t.traverseElements(ct.ComplexContent.Extension.Sequence)
	t.traverseElements(ct.ComplexContent.Extension.Choice)
	t.traverseElements(ct.ComplexContent.Extension.SequenceChoice)
	t.traverseAttributes(ct.SimpleContent.Extension.Attributes)
}

func (t *traverser) traverseAttributes(attrs []*XSDAttribute) {
	for _, attr := range attrs {
		t.traverseAttribute(attr)
	}
}

func (t *traverser) traverseAttribute(attr *XSDAttribute) {
	// Check if we are in ref resolution mode
	if t.tm != refResolution {
		return
	}

	if attr.Ref != "" {
		refAttr := t.getGlobalAttribute(attr.Ref)
		if refAttr != nil && refAttr.Ref == "" {
			t.traverseAttribute(refAttr)
			attr.Name = refAttr.Name
			attr.Type = refAttr.Type
			if attr.Fixed == "" {
				attr.Fixed = refAttr.Fixed
			}
		}
	} else if attr.Type == "" {
		if attr.SimpleType != nil {
			t.traverseSimpleType(attr.SimpleType)
			attr.Type = attr.SimpleType.Restriction.Base
		}
	}
}

func (t *traverser) getGlobalAttribute(name string) *XSDAttribute {
	ref := t.qname(name)

	for _, schema := range t.all {
		if schema.TargetNamespace == ref.Space {
			for _, attr := range schema.Attributes {
				if attr.Name == ref.Local {
					return attr
				}
			}
		}
	}

	return nil
}

func (t *traverser) getGlobalElement(name string) *XSDElement {
	ref := t.qname(name)

	for _, schema := range t.all {
		if schema.TargetNamespace == ref.Space {
			for _, elm := range schema.Elements {
				if elm.Name == ref.Local {
					return elm
				}
			}
		}
	}

	return nil
}

// qname resolves QName into xml.Name.
func (t *traverser) qname(name string) (qname xml.Name) {
	x := strings.SplitN(name, ":", 2)
	if len(x) == 1 {
		qname.Local = x[0]
	} else {
		qname.Local = x[1]
		qname.Space = x[0]
		if ns, ok := t.c.Xmlns[qname.Space]; ok {
			qname.Space = ns
		}
	}

	return qname
}
