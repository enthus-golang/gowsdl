# Namespace Handling Fix Summary

## Problem
After PR #32 was merged, users reported `java.lang.NumberFormatException: For input string: "null"` errors when calling SOAP services. The issue was caused by incorrect namespace handling for schemas where `elementFormDefault` is not specified (defaults to "unqualified").

## Root Cause
When `elementFormDefault="unqualified"` (the default), only top-level elements should be namespace-qualified, while local/child elements should be unqualified. However, when using Go's default namespace declaration (`xmlns="..."`), all child elements inherit the namespace, making them qualified when they shouldn't be.

## Solution
Modified gowsdl to generate operation types with namespace prefixes (e.g., `tns:OperationName`) when `elementFormDefault` is unqualified. This prevents child elements from inheriting the namespace.

## Changes Made

### 1. Enhanced SOAP Envelope (`soap/namespace_envelope.go`)
- Created `EnhancedSOAPEnvelope` struct that supports target namespace declarations
- Added `NewEnhancedSOAPEnvelope` function to create envelopes with proper namespace handling

### 2. Modified SOAP Client (`soap/soap.go`)
- Updated the `call` method to use `EnhancedSOAPEnvelope` when the request implements `TargetNamespace()` method
- Maintains backward compatibility by using standard envelope when target namespace is not available

### 3. Updated Types Template (`pkg/generator/templates/types.go`)
- Modified XMLName generation to use namespace prefix (`tns:`) when `elementFormDefault` is unqualified
- Added `TargetNamespace()` method generation for operation types
- Keeps full namespace format when `elementFormDefault="qualified"`

### 4. Added Comprehensive Tests
- `namespace_qualification_test.go`: Tests for elementFormDefault behavior
- `namespace_complete_test.go`: End-to-end test with user's WSDL structure
- `soap_namespace_test.go`: Tests for SOAP envelope structure

## Example

### Before (Incorrect)
```xml
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <GOWebService_SendungsErstellung xmlns="https://wsdemo.ax4.com/ws/GeneralOvernight">
      <Versender>2611570</Versender>
      <!-- Child elements inherit namespace - WRONG! -->
    </GOWebService_SendungsErstellung>
  </soap:Body>
</soap:Envelope>
```

### After (Correct)
```xml
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" 
               xmlns:tns="https://wsdemo.ax4.com/ws/GeneralOvernight">
  <soap:Body>
    <tns:GOWebService_SendungsErstellung>
      <Versender>2611570</Versender>
      <!-- Child elements are unqualified - CORRECT! -->
    </tns:GOWebService_SendungsErstellung>
  </soap:Body>
</soap:Envelope>
```

## Usage
The fix is automatic. Generated code will now:
1. Use namespace prefixes for operation elements when appropriate
2. Generate `TargetNamespace()` methods on operation types
3. Create proper SOAP envelopes with namespace declarations

## Backward Compatibility
The changes maintain full backward compatibility:
- Existing code continues to work without modifications
- WSDLs with `elementFormDefault="qualified"` continue to use full namespace format
- Standard SOAP envelope is used when target namespace is not available