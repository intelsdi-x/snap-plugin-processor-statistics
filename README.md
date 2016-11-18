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

You can get the pre-built binaries for your OS and architecture from the plugin's [GitHub Releases](https://github.com/intelsdi-x/snap-plugin-processor-statistics/releases) page. Download the plugin from the latest release and load it into `snapd` (`/opt/snap/plugins` is the default location for Snap packages).

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
This builds the plugin in `./build`

### Configuration and Usage
* Set up the [Snap framework](https://github.com/intelsdi-x/snap#getting-started)

## Documentation
This Snap processor plugin calculates statistics over a sliding window. Currently, the plugin calculates the mean, median, standard deviation, variance, 95th-percentile and 99th-percenticle over a sliding window. 

Note: This Snap processor plugin changes the metric data type to `map[string]float64`. Any Snap publisher plugin used with this plugin should register the data type with gob. 

```
import "encoding/gob"

gob.Register(map[string]float64{})
```

### Examples
Example running psutil plugin, statistics processor, and writing data into a file.

Documentation for Snap collector psutil plugin can be found [here](https://github.com/intelsdi-x/snap-plugin-collector-psutil)

In one terminal window, open the Snap daemon :
```
$ snapd -t 0 -l 1
```
The option "-l 1" it is for setting the debugging log level and "-t 0" is for disabling plugin signing.

In another terminal window:

Download and load collector, processor and publisher plugins
```
$ wget http://snap.ci.snap-telemetry.io/plugins/snap-plugin-collector-psutil/latest/linux/x86_64/snap-plugin-collector-psutil
$ wget http://snap.ci.snap-telemetry.io/plugins/snap-plugin-processor-statistics/latest/linux/x86_64/snap-plugin-processor-statistics
$ wget http://snap.ci.snap-telemetry.io/plugins/snap-plugin-publisher-file/latest/linux/x86_64/snap-plugin-publisher-file
$ chmod 755 snap-plugin-*
$ snapctl plugin load snap-plugin-collector-psutil
$ snapctl plugin load snap-plugin-publisher-file
$ snapctl plugin load snap-plugin-processor-statistics
```

See available metrics for your system
```
$ snapctl metric list
```

Create a task file. For example, sample-psutil-statistics-task.json:

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
                "/intel/psutil/load/load1": {},
                "/intel/psutil/load/load5": {},
                "/intel/psutil/load/load15": {},
                "/intel/psutil/vm/free": {},
                "/intel/psutil/vm/used": {}
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

Start task:
```
$ snapctl task create -t sample-psutil-statistics-task.json
Using task manifest to create task
Task created
ID: 02dd7ff4-8106-47e9-8b86-70067cd0a850
Name: Task-02dd7ff4-8106-47e9-8b86-70067cd0a850
State: Running
```

See realtime output from `snapctl task watch <task_id>` (CTRL+C to exit)
```
snapctl task watch 02dd7ff4-8106-47e9-8b86-70067cd0a850
```

This data is published to a file `/tmp/published` per task specification

Stop task:
```
$ snapctl task stop 02dd7ff4-8106-47e9-8b86-70067cd0a850
Task stopped:
ID: 02dd7ff4-8106-47e9-8b86-70067cd0a850
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
