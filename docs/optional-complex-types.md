# Optional Complex Types in gowsdl

## Overview

When a WSDL defines complex types with `minOccurs="0"`, gowsdl generates pointer types for these optional fields. This ensures proper XML marshaling behavior where nil pointers are omitted from the output.

## Behavior

### WSDL Definition
```xml
<element name="shippingAddress" minOccurs="0">
    <complexType>
        <sequence>
            <element name="street" type="string"/>
            <element name="city" type="string"/>
        </sequence>
    </complexType>
</element>
```

### Generated Go Code
```go
type OrderType struct {
    // Optional field (minOccurs="0") - uses pointer
    ShippingAddress *struct {
        Street string `xml:"street,omitempty"`
        City   string `xml:"city,omitempty"`
    } `xml:"shippingAddress,omitempty"`
    
    // Required field (minOccurs="1" or default) - uses value
    BillingAddress struct {
        Street string `xml:"street,omitempty"`
        City   string `xml:"city,omitempty"`
    } `xml:"billingAddress,omitempty"`
}
```

## XML Marshaling Behavior

### With nil pointer (field omitted)
```go
order := OrderType{
    // ShippingAddress is nil
}
// XML output: <order></order>
// No empty <shippingAddress/> element
```

### With non-nil pointer (field included)
```go
order := OrderType{
    ShippingAddress: &struct{
        Street string `xml:"street,omitempty"`
        City   string `xml:"city,omitempty"`
    }{
        Street: "123 Main St",
        City:   "New York",
    },
}
// XML output: <order><shippingAddress><street>123 Main St</street><city>New York</city></shippingAddress></order>
```

## Important: Nil Safety

**When accessing optional fields, always check for nil:**

```go
// WRONG - will panic if ShippingAddress is nil
street := order.ShippingAddress.Street

// CORRECT - check for nil first
if order.ShippingAddress != nil {
    street := order.ShippingAddress.Street
}

// Or use a helper pattern
func (o *OrderType) GetShippingAddressStreet() string {
    if o.ShippingAddress != nil {
        return o.ShippingAddress.Street
    }
    return ""
}
```

## Rationale

This approach follows Go idioms where:
- Optional values are represented by pointers
- The `omitempty` tag works correctly with nil pointers
- It matches how other Go serialization libraries handle optional fields
- No custom MarshalXML methods are needed

## Migration Guide

If you have existing code that accesses optional fields without nil checks, you'll need to update it:

### Before
```go
if order.ShippingAddress.Street != "" {
    // Process shipping address
}
```

### After
```go
if order.ShippingAddress != nil && order.ShippingAddress.Street != "" {
    // Process shipping address
}
```