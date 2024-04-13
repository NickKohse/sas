#!/usr/bin/env ruby
require 'net/http'
require 'tempfile'
require 'digest'

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
        print_colour(:red, "FAIL! Expected response to contain:\n #{expected_result[0..50]}...\n got:\n #{res.body[0..50]}...")
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

run_case("Health Check", "GET", "http://localhost:1997/health", "Uptime", 200)

file = Tempfile.new('test')
str = Time.now.to_s
file_sha256 = Digest::SHA256.hexdigest str
file.write(Time.now)

large_file = Tempfile.new('largetest')
large_str = 'a' * 10_000_000
large_file.write(large_str)
large_file_sha256 = Digest::SHA256.hexdigest large_str

file.close
large_file.close

run_case("File Upload", "POST", "http://localhost:1997/artifact", "Successfully Uploaded File", 201, {artifact: file.path})
run_case("File Download", "GET", "http://localhost:1997/artifact", str, 200, {artifact: File.basename(file.path)})
run_case("Get Metadata", "GET", "http://localhost:1997/metadata", "\"Sha256\":\"#{file_sha256}\"", 200, {artifact: File.basename(file.path)})
run_case("Get Checksum", "GET", "http://localhost:1997/checksum", file_sha256, 200, {artifact: File.basename(file.path)})
run_case("Search", "GET", "http://localhost:1997/search?q=test", "test", 200)
run_case("Delete File", "DELETE", "http://localhost:1997/artifact", "Successfully Deleted ", 200, {artifact: File.basename(file.path)})
run_case("File Download After it's Deleted", "GET", "http://localhost:1997/artifact", "", 404, {artifact: File.basename(file.path)})
run_case("Get Metadata After it's Deleted", "GET", "http://localhost:1997/metadata", "", 404, {artifact: File.basename(file.path)})
run_case("Get Checksum After it's Deleted", "GET", "http://localhost:1997/checksum", "", 404, {artifact: File.basename(file.path)})
run_case("Invalid File Upload", "POST", "http://localhost:1997/artifact", "Unable to process artifact", 400, {artifact: ''})
run_case("File Download - No Artifact Specified", "GET", "http://localhost:1997/artifact", "", 400)
run_case("Get Metadata - No Artifact Specifiedd", "GET", "http://localhost:1997/metadata", "", 400)
run_case("Get Checksum - No Artifact Specified", "GET", "http://localhost:1997/checksum", "", 400)

run_case("Large File Upload", "POST", "http://localhost:1997/artifact", "Successfully Uploaded File", 201, {artifact: large_file.path})
run_case("Large File Download", "GET", "http://localhost:1997/artifact", large_str, 200, {artifact: File.basename(large_file.path)})
run_case("Large File Get Metadata", "GET", "http://localhost:1997/metadata", "\"Sha256\":\"#{large_file_sha256}\"", 200, {artifact: File.basename(large_file.path)})
run_case("Large File Get Checksum", "GET", "http://localhost:1997/checksum", large_file_sha256, 200, {artifact: File.basename(large_file.path)})
run_case("Delete Large File", "DELETE", "http://localhost:1997/artifact", "Successfully Deleted ", 200, {artifact: File.basename(large_file.path)})

elapsed = Time.now - start

print_colour(:orange, "Run of #{$cases} cases complete. #{$passes}/#{$cases} passed. Time elapsed #{elapsed} seconds.\n")
