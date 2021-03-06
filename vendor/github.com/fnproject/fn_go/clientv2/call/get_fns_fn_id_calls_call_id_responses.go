// Code generated by go-swagger; DO NOT EDIT.

package call

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	modelsv2 "github.com/fnproject/fn_go/modelsv2"
)

// GetFnsFnIDCallsCallIDReader is a Reader for the GetFnsFnIDCallsCallID structure.
type GetFnsFnIDCallsCallIDReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetFnsFnIDCallsCallIDReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewGetFnsFnIDCallsCallIDOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 404:
		result := NewGetFnsFnIDCallsCallIDNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewGetFnsFnIDCallsCallIDOK creates a GetFnsFnIDCallsCallIDOK with default headers values
func NewGetFnsFnIDCallsCallIDOK() *GetFnsFnIDCallsCallIDOK {
	return &GetFnsFnIDCallsCallIDOK{}
}

/*GetFnsFnIDCallsCallIDOK handles this case with default header values.

Call found.
*/
type GetFnsFnIDCallsCallIDOK struct {
	Payload *modelsv2.Call
}

func (o *GetFnsFnIDCallsCallIDOK) Error() string {
	return fmt.Sprintf("[GET /fns/{fnID}/calls/{callID}][%d] getFnsFnIdCallsCallIdOK  %+v", 200, o.Payload)
}

func (o *GetFnsFnIDCallsCallIDOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(modelsv2.Call)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetFnsFnIDCallsCallIDNotFound creates a GetFnsFnIDCallsCallIDNotFound with default headers values
func NewGetFnsFnIDCallsCallIDNotFound() *GetFnsFnIDCallsCallIDNotFound {
	return &GetFnsFnIDCallsCallIDNotFound{}
}

/*GetFnsFnIDCallsCallIDNotFound handles this case with default header values.

Call not found.
*/
type GetFnsFnIDCallsCallIDNotFound struct {
	Payload *modelsv2.Error
}

func (o *GetFnsFnIDCallsCallIDNotFound) Error() string {
	return fmt.Sprintf("[GET /fns/{fnID}/calls/{callID}][%d] getFnsFnIdCallsCallIdNotFound  %+v", 404, o.Payload)
}

func (o *GetFnsFnIDCallsCallIDNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(modelsv2.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
