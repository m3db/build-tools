// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package testdata

import (
	"context"
	"fmt"
	"time"

	"github.com/m3db/m3coordinator/models"
	"github.com/m3db/m3coordinator/services/m3coordinator/config"
	"github.com/m3db/m3coordinator/services/m3coordinator/httpd"
	"github.com/m3db/m3coordinator/util/logging"

	"github.com/m3db/m3db/client"
	xtime "github.com/m3db/m3x/time"

	"github.com/m3db/m3db/digest"
	"go.uber.org/zap"
	"gopkg.in/alecthomas/kingpin.v2"
)

func test3() {
	fmt.Println("import fmt", time.Now())
	var _ = config.Configuration{}
	var _ = models.Tags{}
	var _ = digest.NewBuffer()
	var _ = httpd.Handler{}
	var _ = logging.WithContext(context.TODO())
	var _ = client.NewOptions()
	var _ = xtime.Millisecond
	var _ = kingpin.Arg("", "")
	var _ = zap.String("address", "")
}
