# Snap Processor Plugin - Statistics
Snap plugin intended to process data and return statistics over a sliding window.

1. [Getting Started](#getting-started)
  * [System Requirements](#system-requirements)
  * [Installation](#installation)
  * [Configuration and Usage](configuration-and-usage)
2. [Documentation](#documentation)
  * [Examples](#examples)
  * [Roadmap](#roadmap)
3. [Community Support](#community-support)
4. [Contributing](#contributing)
5. [License](#license)
6. [Acknowledgements](#acknowledgements)


## Getting Started
### System Requirements
* Linux/amd64

### Installation
#### Download plugin binary: 

You can get the pre-built binaries for your OS and architecture at Snap's [Github Releases](https://github.com/intelsdi-x/snap/releases) page.

#### To build the plugin binary:
Fork https://github.com/intelsdi-x/snap-plugin-processor-statistics

Clone repo into `$GOPATH/src/github/intelsdi-x/`:
```
$ git clone https://github.com/<yourGithubID>/snap-plugin-processor-statistics
```
Build the plugin by running make in repo:
```
$ make
```
This builds the plugin in `/build/rootfs`

### Configuration and Usage
* Set up the [snap framework](https://github.com/intelsdi-x/snap/blob/master/README.md#getting-started)
* Ensure `$SNAP_PATH` is exported
`export SNAP_PATH=$GOPATH/src/github.com/intelsdi-x/snap/build`

## Documentation
This Snap processor plugin calculates statistics over a sliding window. Currently, the plugin calculates the mean, median, standard deviation, variance, 95th-percentile and 99th-percenticle over a sliding window. 

Note: This Snap processor plugin changes the metric data type to `map[string]float64`. Any Snap publisher plugin used with this plugin should register the data type with gob. 

```
import "encoding/gob"

gob.Register(map[string]float64{})
```

### Examples
Creating a task manifest file. 
```
{
    "version": 1,
    "schedule": {
        "type": "simple",
        "interval": "1s"
    },
    "workflow": {
        "collect": {
            "metrics": {
                "/intel/mock/foo": {},
                "/intel/mock/bar": {},
                "/intel/mock/*/baz": {}
            },
            "config": {
                "/intel/mock": {
                    "user": "root",
                    "password": "secret"
                }
            },
            "process": [
                {
                    "plugin_name": "statistics",
		    "config":
    			{
	    			"SlidingWindowLength": 15
			},		
                    "process": null,
                    "publish": [
                        {
                            "plugin_name": "file",
                            "config": {
                                "file": "/tmp/published"
                            }
                        }
                    ]
                }
            ]
        }
    }
}
```

### Roadmap
There isn't a current roadmap for this plugin, but it is in active development. As we launch this plugin, we do not have any outstanding requirements for the next release.

If you have a feature request, please add it as an [issue](https://github.com/intelsdi-x/snap-plugin-processor-statistics/issues) and feel free to then submit a [pull request](https://github.com/intelsdi-x/snap-plugin-processor-statistics/pulls).

## Community Support
This repository is one of **many** plugins in **Snap**, the open telemetry framework. See the full project at http://github.com/intelsdi-x/snap. To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support).

## Contributing
We love contributions!

There's more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

And **thank you!** Your contribution, through code and participation, is incredibly important to us.

## License
[Snap](http://github.com:intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).

## Acknowledgements

* Author: [Balaji Subramaniam](https://github.com/balajismaniam)

And **thank you!** Your participation and contribution through code is important to us.
