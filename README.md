# Azure Application Insights Tracing

Wrapper over application insights package that allows for richer telemetry and
w3c tracing support by exposing a more simplified interface

## Installation
1. Install module
```go
go get github.com/BetaLixT/appInsightsTrace
```
2. Import
```go
import "github.com/BetaLixT/appInsightsTrace"
```

## Usage
Create an application insights resource on Azure, and note down the
instrumentation key

Construct a new AppInsightsCore instance using one of the existing constructors
read the documentations on the constructors to learn more about them, for a
quick start the NewBasic constructor is recommended but the NewAppInsightsCore
with a custom implementation of ITraceExtractor is recommended for production 
use

In this example we use the NewBasic, provide the above mentioned
instrumentationKey and the service name would be the name of the service/project
you will be using this package in, this will be used as an application name when
viewing the application map on Azure App Insights
```go
tracer := appInsightsTrace.NewBasic(instrumentationKey, "WeatherService")
```

Once the struct has been created you can utilize one of the Trace functions to
start tracing events, all telemetery you create with these functions will be
batched and sent to application insights. Please read the documentation on these
functions for more details, when using NewBasic (since this uses the default
ITraceExtractor) it's recommended to use the WithIds functions so you can
manually pass down the tracing information (instead of relying on context)
```go
tracer.TraceRequestWithIds(
  "3251235485234561",
  "35623452",
  "35123514",
  "GET",
  "api/v1/weather",
  200,
  7000,
  "127.0.0.1",
  "Chrome",
  startTime,
  time.Now(),
  map[string]string{},
)
```

## Recommendation
It's recommended that you create an implementation of ITraceExtractor according
to how you handle w3c tracing within your service.

An example usage would be to have a request middleware that will check for
trace-parent in the header or generate one if it doesn't exist, this
trace-parent can then be parsed and along with the current request(span) id be
injected into the context. You can then create an implementation of
ITraceExtractor to extract said trace information from the context
