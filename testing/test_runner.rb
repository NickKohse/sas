#!/usr/local/bin/ruby
require 'net/http'
require 'tempfile'

# TODO opt parser to choose if were testing a `go run` build or a compiled build

# Spin up server according to option chosen

$cases = 0
$passes= 0

def run_case(title, method, url, expected_result, expected_status_code, params = {})
    $cases += 1
    print_colour(:orange, "Running case: '#{title}'... ")
    uri = URI(url)
    # case statemetn and its code could use some cleanup, redundant and makes weird use of param argument
    case method
    when 'GET'
        req = Net::HTTP::Get.new(uri)
        if params.has_key? :artifact
            req.set_form_data(params)
            form_data = [['artifact', params[:artifact]]]
            req.set_form form_data, 'multipart/form-data'
            res = Net::HTTP.start(uri.hostname, uri.port) do |http|
              http.request(req)
            end
        end
    when 'POST'
        req = Net::HTTP::Post.new(uri)
        form_data = []
        if params[:artifact] == ''
            form_data = [['artifact', '']]
        else
            form_data = [['artifact', File.open(params[:artifact])]]
        end
        req.set_form form_data, 'multipart/form-data'
    when 'DELETE'
        req = Net::HTTP::Delete.new(uri)
        if params.has_key? :artifact
            req.set_form_data(params)
            form_data = [['artifact', params[:artifact]]]
            req.set_form form_data, 'multipart/form-data'
        end
    end

    res = Net::HTTP.start(uri.hostname, uri.port) do |http|
        http.request(req)
    end
    failure = false
    if res.code.to_i != expected_status_code
        print_colour(:red, "FAIL! Expected status code #{expected_status_code} got #{res.code} ")
        failure = true
    end
    unless res.body.include? expected_result
        print_colour(:red, "FAIL! Expected response to contain:\n #{expected_result}\n got:\n #{res.body}")
        failure = true
    end
    return if failure
    $passes += 1
    print_colour(:green, "PASS! Test case: '#{title}' passed\n")
end

def print_colour(colour, str)
    code = { orange: 33, green: 32, red: 31 }[colour]
    print "\e[#{code}m#{str}\e[0m"
end

start = Time.now

run_case("Health Check", "GET", "http://localhost:1997/health", "Uptime:", 200)

file = Tempfile.new('test')
file.write(Time.now)
file.close
run_case("File Upload", "POST", "http://localhost:1997/artifact", "Successfully Uploaded File", 201, {artifact: file.path})
run_case("File Download", "GET", "http://localhost:1997/artifact", "", 200, {artifact: File.basename(file.path)}) # Return file here and ensure its what we expect
run_case("Get Metadata", "GET", "http://localhost:1997/metadata", "CreateTime", 200, {artifact: File.basename(file.path)}) # Confirm validity of metadata
run_case("Get Checksum", "GET", "http://localhost:1997/checksum", "", 200, {artifact: File.basename(file.path)}) #ditto
run_case("Search", "GET", "http://localhost:1997/search?q=test", "test", 200)
run_case("Delete File", "DELETE", "http://localhost:1997/artifact", "Successfully deleted ", 200, {artifact: File.basename(file.path)})
run_case("File Download After it's Deleted", "GET", "http://localhost:1997/artifact", "", 404, {artifact: File.basename(file.path)})
run_case("Get Metadata After it's Deleted", "GET", "http://localhost:1997/metadata", "", 404, {artifact: File.basename(file.path)})
run_case("Get Checksum After it's Deleted", "GET", "http://localhost:1997/checksum", "", 404, {artifact: File.basename(file.path)})
run_case("Invalid File Upload", "POST", "http://localhost:1997/artifact", "Unable to process artifact", 400, {artifact: ''})
run_case("File Download - No Artifact Specified", "GET", "http://localhost:1997/artifact", "", 400)
run_case("Get Metadata - No Artifact Specifiedd", "GET", "http://localhost:1997/metadata", "", 400)
run_case("Get Checksum - No Artifact Specified", "GET", "http://localhost:1997/checksum", "", 400)

elapsed = Time.now - start

print_colour(:orange, "Run of #{$cases} cases complete. #{$passes}/#{$cases} passed. Time elapsed #{elapsed} seconds.\n")
