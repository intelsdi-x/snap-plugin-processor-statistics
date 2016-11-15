require 'json'
require 'pathname'
require 'yaml'
require 'rspec/retry'
require 'dockerspec/serverspec'

begin
  require 'pry'
rescue LoadError
end

module SnapUtils
  def sh(arg)
    c = command(arg)
    puts c.stderr
    puts c.stdout
  end

  def cmd_with_retry(arg, opt={ :timeout => 30 })
    cmd = command(arg)
    while cmd.exit_status != 0 or cmd.stdout == '' and opt[:timeout] > 0
      sleep 5
      opt[:timeout] -= 5
      cmd = command(arg)
    end
    return cmd
  end

  def curl_json_api(url)
    output = cmd_with_retry("curl #{url}").stdout
    if output.size > 0
      JSON.parse(output)
    else
      {}
    end
  end

  def load_json(file)
    file = File.expand_path file
    raise ArgumentError, "Invalid json file path: #{file}" unless File.exist? file
    JSON.parse File.read file
  end

  def load_yaml(file)
    file = File.expand_path file
    raise ArgumentError, "Invalid json file path: #{file}" unless File.exist? file
    YAML.load_file file
  end

  def self.examples
    Pathname.new(File.expand_path(File.join(__FILE__,'../../../examples')))
  end

  def self.tasks
    Dir.glob("#{examples}/tasks/*.yml").collect{|f| File.basename f}
  end
end

RSpec.configure do |c|
  c.formatter = 'documentation'
  c.mock_framework = :rspec
  c.verbose_retry = true
  c.order = 'default'
  c.include SnapUtils
end
