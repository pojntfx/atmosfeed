// Code generated by scale-signature v0.3.20, DO NOT EDIT.
// output: signature

package signature

import (
	"github.com/loopholelabs/polyglot"
	interfaces "github.com/loopholelabs/scale-signature-interfaces"
)

const hash = "4a8a84f12b689b5b500cd44fc143e32a7432a548a8ba913a8deef52fa3a0ce12"

var _ interfaces.Signature = (*Signature)(nil)

// Signature is the host representation of the signature
//
// Users should not use this type directly, but instead pass the New() function
// to the Scale Runtime
type Signature struct {
	Context *Context
	buf     *polyglot.Buffer
}

// New returns a new signature and tells the Scale Runtime how to use it
//
// This function should be passed into the scale runtime config as an argument
func New() *Signature {
	return &Signature{
		Context: NewContext(),
		buf:     polyglot.NewBuffer(),
	}
}

// Read reads the context from the given byte slice and returns an error if one occurred
//
// This method is meant to be used by the Scale Runtime to deserialize the Signature
func (x *Signature) Read(b []byte) error {
	var err error
	x.Context, err = DecodeContext(x.Context, b)
	return err
}

// Write writes the signature into a byte slice and returns it
//
// This method is meant to be used by the Scale Runtime to serialize the Signature
func (x *Signature) Write() []byte {
	x.buf.Reset()
	x.Context.Encode(x.buf)
	return x.buf.Bytes()
}

// Error writes the signature into a byte slice and returns it
//
// This method is meant to be used by the Scale Runtime to return an error
func (x *Signature) Error(err error) []byte {
	x.buf.Reset()
	polyglot.Encoder(x.buf).Error(err)
	return x.buf.Bytes()
}

// Hash returns the hash of the signature
//
// This method is meant to be used by the Scale Runtime to validate Signature and Function compatibility
func (x *Signature) Hash() string {
	return hash
}
