# Inline Complex Types Solution

## Overview

Starting from version [current], gowsdl automatically generates named types for all inline complex types. This solves the issue of anonymous struct pointers being difficult to use in Go.

## The Solution

When WSDL defines inline complex types, gowsdl now:
1. Generates named types for all inline complex types
2. Uses pointers to these named types for optional fields (`minOccurs="0"`)
3. Generates descriptive type names based on the parent type hierarchy

Example WSDL:
```xml
<complexType name="OrderType">
    <sequence>
        <element name="shippingAddress" minOccurs="0">
            <complexType>
                <sequence>
                    <element name="street" type="string"/>
                    <element name="city" type="string"/>
                </sequence>
            </complexType>
        </element>
    </sequence>
</complexType>
```

Generated Go code:
```go
type OrderType struct {
    ShippingAddress *OrderTypeShippingAddressType `xml:"shippingAddress,omitempty"`
}

type OrderTypeShippingAddressType struct {
    Street string `xml:"street,omitempty"`
    City   string `xml:"city,omitempty"`
}
```

## Benefits

1. **Type Safety**: All types are named and can be referenced
2. **Easy to Use**: No need for complex type assertions or anonymous struct literals
3. **Idiomatic Go**: Follows Go best practices for optional fields (using pointers)
4. **Proper XML Omission**: Nil pointers are omitted from XML output
5. **Empty Struct Handling**: Generated types include MarshalXML methods to prevent empty complex types from being marshaled

## Usage Example

```go
// Creating an instance with optional field
order := &OrderType{
    ShippingAddress: &OrderTypeShippingAddressType{
        Street: "123 Main St",
        City:   "NYC",
    },
}

// Optional field can be nil
order2 := &OrderType{
    // ShippingAddress is nil, will be omitted from XML
}

// Accessing the optional field
if order.ShippingAddress != nil {
    fmt.Println("Street:", order.ShippingAddress.Street)
}
```

## Naming Convention

Generated type names follow this pattern:
- `{ParentType}{FieldName}Type`
- For nested types: `{ParentType}{FieldName}Type{NestedFieldName}Type`

Examples:
- `OrderType.shippingAddress` → `OrderTypeShippingAddressType`
- `CompanyType.headquarters.address` → `CompanyTypeHeadquartersTypeAddressType`

## Migration from Previous Versions

If you were using workarounds for anonymous structs, you can now:
1. Remove helper type definitions
2. Use the generated named types directly
3. Update any type assertions to use the new type names

## Backward Compatibility

This is a breaking change if you were:
- Using anonymous struct literals directly
- Using type assertions with anonymous struct types

To migrate, replace anonymous struct references with the generated named types.

## Empty Struct Handling

Starting from this version, all generated inline complex types include a custom `MarshalXML` method that checks if all fields are at their zero values. If so, the entire element is omitted from the XML output.

This prevents issues with SOAP servers that cannot handle empty complex elements, such as:
```xml
<!-- This causes errors on some servers -->
<Ansprechpartner>
    <Telefon></Telefon>
</Ansprechpartner>

<!-- With MarshalXML, the entire element is omitted if empty -->
```

Example generated code:
```go
type ContactTypePhoneInfoType struct {
    Phone  string `xml:"phone,omitempty"`
    Mobile string `xml:"mobile,omitempty"`
}

func (t ContactTypePhoneInfoType) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
    // Check if all fields are at their zero value
    if t.Phone == "" && t.Mobile == "" {
        // Skip marshaling this empty struct
        return nil
    }
    // Use type alias to avoid infinite recursion
    type alias ContactTypePhoneInfoType
    return e.EncodeElement(alias(t), start)
}
```

This ensures that even if you create a pointer to an empty struct, it won't produce invalid XML:
```go
contact := ContactType{
    Name: "John Doe",
    PhoneInfo: &ContactTypePhoneInfoType{}, // Empty struct
}
// PhoneInfo will be completely omitted from XML output
```