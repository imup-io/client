package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/m-lab/ndt7-client-go/spec"
	"github.com/olekukonko/tablewriter"
)

// EmitterOutput is a human readable emitter. It emits the events generated
// by running a ndt7 test as pleasant stdout messages.
type EmitterOutput struct {
	out io.Writer
}

// NewEmitterOutput returns a new human readable emitter using the
// specified writer.
func NewEmitterOutput(w io.Writer) *EmitterOutput {
	return &EmitterOutput{w}
}

// Started handles the start event
func (e EmitterOutput) Started(test spec.TestKind) error {
	_, err := fmt.Fprintf(e.out, "\nstarting %s\n", test)
	return err
}

// Failed handles an error event
func (e EmitterOutput) Failed(test spec.TestKind, err error) error {
	_, failure := fmt.Fprintf(e.out, "\n%s failed: %s\n", test, err.Error())
	return failure
}

// Connected handles the connected event
func (e EmitterOutput) Connected(test spec.TestKind, fqdn string) error {
	_, err := fmt.Fprintf(e.out, "\n%s in progress with %s\n", test, fqdn)
	return err
}

// SpeedEvent handles the emitter output
func (e EmitterOutput) SpeedEvent(m *spec.Measurement) error {
	// The specification recommends that we show application level
	// measurements. Let's just do that in interactive mode. To this
	// end, we ignore any measurement coming from the server.
	if m.Origin != spec.OriginClient {
		return nil
	}
	if m.AppInfo == nil || m.AppInfo.ElapsedTime <= 0 {
		return errors.New("missing m.AppInfo or invalid m.AppInfo.ElapsedTime")
	}
	elapsed := float64(m.AppInfo.ElapsedTime) / 1e06
	v := (8.0 * float64(m.AppInfo.NumBytes)) / elapsed / (1000.0 * 1000.0)
	_, err := fmt.Fprintf(e.out, "\r%7.1f Mbit/s", v)
	return err
}

// Completed posts a complete event notification
func (e EmitterOutput) Completed(test spec.TestKind) error {
	_, err := fmt.Fprintf(e.out, "\n%s: completed\n", test)

	return err
}

// Summary is a tabledized summary of the test activity
func (e EmitterOutput) Summary(t *speedTestData) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Type", "Measurement", "Value"})

	table.Append([]string{"Download", "Download", fmt.Sprintf("%.2f Mbit/s", t.DownloadMbps)})
	table.Append([]string{"Download", "DownloadedBytes", fmt.Sprintf("%.2f Bytes", t.DownloadedBytes)})
	table.Append([]string{"Download", "DownloadRetrans", fmt.Sprintf("%.4f %%", t.DownloadRetrans)})
	table.Append([]string{"Download", "DownloadMinRTT", fmt.Sprintf("%.4f ms", t.DownloadMinRtt)})
	table.Append([]string{"Download", "DownloadRTTVar", fmt.Sprintf("%.4f ms", t.DownloadRTTVar)})

	table.Append([]string{"Uplaod", "Upload", fmt.Sprintf("%.2f Mbit/s", t.UploadMbps)})
	table.Append([]string{"Uplaod", "UploadedBytes", fmt.Sprintf("%.2f Bytes", t.UploadedBytes)})
	table.Append([]string{"Uplaod", "UploadRetrans", fmt.Sprintf("%.4f %%", t.UploadRetrans)})
	table.Append([]string{"Uplaod", "UploadMinRTT", fmt.Sprintf("%.4f ms", t.UploadMinRTT)})
	table.Append([]string{"Uplaod", "UploadRTTVar", fmt.Sprintf("%.4f ms", t.UploadRTTVar)})
	for k, v := range t.Metadata {
		table.Append([]string{"metadata", k, v})
	}

	table.Render()
}

func (e EmitterOutput) testRunner(ctx context.Context, fqdn string, kind spec.TestKind, start startFunc) error {
	ch, err := start(ctx)
	if err != nil {
		e.Failed(kind, err)
		return err
	}

	err = e.Started(kind)
	if err != nil {
		e.Failed(kind, err)
		return err
	}

	err = e.Connected(kind, fqdn)
	if err != nil {
		e.Failed(kind, err)
		return err
	}

	errChan := make(chan error)
	for event := range ch {
		func(m *spec.Measurement) {
			err := e.SpeedEvent(&event)
			if err != nil {
				errChan <- err
			}
		}(&event)
	}

	close(errChan)

	var errs error
	for err = range errChan {
		if err != nil {
			errs = fmt.Errorf("%v", err)
		}
	}
	if errs != nil {
		return errs
	}

	err = e.Completed(kind)
	if err != nil {
		e.Failed(kind, err)
		return err
	}

	return nil
}
