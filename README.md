
# DISCONTINUATION OF PROJECT 

**This project will no longer be maintained by Intel.  Intel will not provide or guarantee development of or support for this project, including but not limited to, maintenance, bug fixes, new releases or updates.  Patches to this project are no longer accepted by Intel. If you have an ongoing need to use this project, are interested in independently developing it, or would like to maintain patches for the community, please create your own fork of the project.**


# Snap Processor Plugin - Statistics
Snap plugin intended to process data and return statistics over a sliding window.

1. [Getting Started](#getting-started)
  * [System Requirements](#system-requirements)
  * [Installation](#installation)
  * [Configuration and Usage](#configuration-and-usage)
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
* OS X

### Installation
#### Download plugin binary: 

You can get the pre-built binaries for your OS and architecture from the plugin's [GitHub Releases](https://github.com/intelsdi-x/snap-plugin-processor-statistics/releases) page. Download the plugin from the latest release and load it into `snapteld` (`/opt/snap/plugins` is the default location for Snap packages).

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

Unit testing:
```
$ make test-small
```

### Configuration and Usage
* Set up the [Snap framework](https://github.com/intelsdi-x/snap#getting-started)

## Documentation
This Snap processor plugin calculates statistics over a sliding window. To illustrate the sliding window concept, let's say we have a sliding window of 10 over 100 points long, if the sliding factor is 1 and we want to calculate the statistic "average", it will take the average from 0-10, 1-11, 2-12, 3-13 and so on. That takes the 100 data points and averages to 99 points.
Sliding window can be used to smooth out a waveform thereby reducing noise. It is very useful in handling noisey data and also in live streaming of data.
In regular windowing, it takes the average of data points 0-10, then 11-20, then 21-30 and so on. So there will be 10 points instead of 100.
So in regular windowing, we have two variables, the size of the dataset, and the size of the window. Since sliding windows "overlap" the previous and following windows, we have the total size, window size and sliding factor.  		

Here's an example of the sliding window length concept and also the list of statistics calculated by the processor statistics plugin along with their descriptions-
[Sliding Window.pdf](https://github.com/intelsdi-x/snap-plugin-processor-statistics/files/599298/Sliding.Window.pdf)

The default values of sliding factor is 1 and the interval is 1s. Sliding window length default is 100.		
		 
### Examples
Example running psutil plugin, statistics processor, and writing data into a file.

Documentation for Snap collector psutil plugin can be found [here](https://github.com/intelsdi-x/snap-plugin-collector-psutil)

In one terminal window, open the Snap daemon :
```
$ snapteld -t 0 -l 1
```
The option "-l 1" it is for setting the debugging log level and "-t 0" is for disabling plugin signing.

In another terminal window:

Download and load collector, processor and publisher plugins
```
$ wget http://snap.ci.snap-telemetry.io/plugins/snap-plugin-collector-psutil/latest/linux/x86_64/snap-plugin-collector-psutil
$ wget http://snap.ci.snap-telemetry.io/plugins/snap-plugin-processor-statistics/latest/linux/x86_64/snap-plugin-processor-statistics
$ wget http://snap.ci.snap-telemetry.io/plugins/snap-plugin-publisher-file/latest/linux/x86_64/snap-plugin-publisher-file
$ chmod 755 snap-plugin-*
$ snaptel plugin load snap-plugin-collector-psutil
$ snaptel plugin load snap-plugin-publisher-file
$ snaptel plugin load snap-plugin-processor-statistics
```

See available metrics for your system
```
$ snaptel metric list
```

Create a task file. For example, psutil-statistics-file.json:

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
          "config": {
             "SlidingWindowLength": 5,
             "SlidingFactor": 2
          },
          "process": null,
          "publish": [
            {
              "plugin_name": "file",
              "config": {
                "file": "/tmp/published_statistics.log"
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
$ snaptel task create -t psutil-statistics-file.json
Using task manifest to create task
Task created
ID: 02dd7ff4-8106-47e9-8b86-70067cd0a850
Name: Task-02dd7ff4-8106-47e9-8b86-70067cd0a850
State: Running
```

See realtime output from `snaptel task watch <task_id>` (CTRL+C to exit)
```
snaptel task watch 02dd7ff4-8106-47e9-8b86-70067cd0a850
```

This data is published to a file `/tmp/published_statistics.log` per task specification

Stop task:
```
$ snaptel task stop 02dd7ff4-8106-47e9-8b86-70067cd0a850
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
[Snap](http://github.com/intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).

## Acknowledgements

* Authors: [Rashmi Gottipati](https://github.com/rashmigottipati),
           [Balaji Subramaniam](https://github.com/balajismaniam)

And **thank you!** Your participation and contribution through code is important to us.
