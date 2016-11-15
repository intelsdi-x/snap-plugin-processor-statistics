require 'spec_helper'

describe docker_build(template: SnapUtils.examples/"Dockerfile.erb", log_level: :info) do
  describe docker_run(described_image, :env => {'SNAP_VERSION'=>ENV['SNAP_VERSION']}, :wait => 10) do
    describe file("/opt/snap/bin/snapd"), :retry => 3, :retry_wait => 5 do
      it { should be_file }
      it { should be_executable }
    end

    describe file("/opt/snap/bin/snapctl") do
      it { should be_file }
      it { should be_executable }
    end

    describe command("snapd --version") do
      its(:exit_status) { should eq 0 }
      its(:stdout) { should contain /#{ENV['SNAP_VERSION']}/ }
    end if ENV['SNAP_VERSION'] =~ /^\d+.\d+.\d+$/

    # NOTE: using ip instead of localhost due a known issue with alpine:
    # https://github.com/gliderlabs/docker-alpine/issues/8
    snapctl = "snapctl -u 'http://127.0.0.1:8181'"

    describe "snapctl plugin list", :retry => 3, :retry_wait => 10 do

      it {
        cmd=cmd_with_retry("#{snapctl} plugin list")
        expect(cmd.exit_status).to eq 0
        expect(cmd.stdout).to contain /psutil/
        expect(cmd.stdout).to contain /file/
        expect(cmd.stdout).to contain /statistics/
      }
    end

    SnapUtils.tasks.each do |t|
      context "Snap task #{t}" do
        task_id = nil

        describe command("#{snapctl} task create -t /opt/snap/tasks/#{t}") do
          its(:exit_status) { should eq 0 }
          its(:stdout) { should contain /Task created/ }
          it {
            id = subject.stdout.split("\n").find{|l|l=~/^ID:/}
            task_id = $1 if id.match(/^ID: (.*)$/)
            expect(task_id).to_not be_nil
          }
        end
       
        describe "snapctl task list" do
          it {
            # NOTE: The sleep here is to ensure the plugin isn't disabled
            sleep 15
            cmd=command("#{snapctl} task list")
            expect(cmd.exit_status).to eq 0
            expect(cmd.stdout).to contain /Running/
          }
        end

        describe "Metrics in running tasks" do
          it {
            data = curl_json_api("http://127.0.0.1:8181/v1/tasks")
            task = data["body"]["ScheduledTasks"].find{|i| i['id'] == task_id}
            expect(task['id']).to eq task_id
            data = curl_json_api(task['href'])
            collect_metrics = data["body"]["workflow"]["collect"]["metrics"].collect{|k,v| k}

            config = load_yaml(SnapUtils.examples/"tasks/#{t}")
            config_metrics = config['workflow']['collect']['metrics'].collect{|k,v| k}

            config_metrics.each do |m|
              expect(collect_metrics).to include(m)
            end
          }
        end

        # NOTE: can not use the normal describe command(...) since we need to access task_id
        describe "Stop task" do
          it {
            c = command("#{snapctl} task stop #{task_id}")
            expect(c.exit_status).to eq 0
            expect(c.stdout).to match /Task stopped/
          }
        end

        describe "Remove task" do
          it {
            c = command("#{snapctl} task remove #{task_id}")
            expect(c.exit_status).to eq 0
            expect(c.stdout).to match /Task removed/
          }
        end
      end
    end
  end
end