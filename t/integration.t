#!/usr/bin/env perl
use strict;
use warnings;

use Test::More;
use Test::Deep;
use LWP::UserAgent;
use HTTP::Request;
use JSON qw/decode_json encode_json/;
use POSIX qw/strftime/;
use MIME::Base64 qw/encode_base64/;

$ENV{PUBLIC_PORT} = 9077;

my $UA = LWP::UserAgent->new(agent => 'ssg-test/'.strftime("%Y%m%dT%H%M%S", gmtime())."-$$");
my $BASE_URL = "http://127.0.0.1:$ENV{PUBLIC_PORT}";
my ($id, $TOKEN, $AUTH);
sub as_control { $AUTH = ['test-control', 't-c-sekrit']; }
sub as_admin   { $AUTH = ['test-admin',   't-a-sekrit']; }
sub as_agent   { $AUTH = undef; }

sub maybe_json {
	my ($raw) = @_;
	return eval { return decode_json($raw) } or undef;
}

my ($req, $res, $SUCCESS, $STATUS, $RESPONSE);
sub GET {
  my ($url) = @_;
  $url =~ s|^/||;

  $req = HTTP::Request->new(GET => "$BASE_URL/$url")
  	or die "failed to make [GET /$url] request: $!\n";
  $req->authorization_basic(@$AUTH)     if  $AUTH;
  $req->header('X-SSG-Token' => $TOKEN) if !$AUTH && $TOKEN;
  $req->header('Accept' => 'application/json');

  diag $req->as_string if $ENV{DEBUG_HTTP};
  $res = $UA->request($req);
  diag $res->as_string if $ENV{DEBUG_HTTP};
  ($SUCCESS, $STATUS, $RESPONSE) = ($res->is_success, $res->code, maybe_json($res->decoded_content));
}

sub POST {
  my ($url, $payload) = @_;
  $url =~ s|^/||;

  $req = HTTP::Request->new(POST => "$BASE_URL/$url")
  	or die "failed to make [POST /$url] request: $!\n";
  $req->authorization_basic(@$AUTH)     if  $AUTH;
  $req->header('X-SSG-Token' => $TOKEN) if !$AUTH && $TOKEN;
  $req->header('Accept'       => 'application/json');
  $req->header('Content-Type' => 'application/json') if $payload;
  $req->content(encode_json($payload))               if $payload;

  diag $req->as_string if $ENV{DEBUG_HTTP};
  $res = $UA->request($req);
  diag $res->as_string if $ENV{DEBUG_HTTP};
  ($SUCCESS, $STATUS, $RESPONSE) = ($res->is_success, $res->code, maybe_json($res->decoded_content));
}

diag "setting up docker-compose integration environment...\n";
system('t/setup');
is $?, 0, 't/setup should exit zero (success)'
  or done_testing;

as_agent;
POST '/download', { path => 'some/path/some/where' };
ok !$SUCCESS, "attempting to create a download as the agent should fail"
	or diag $res->as_string;
POST '/upload', { path => 'some/path/some/where' };
ok !$SUCCESS, "attempting to create an upload as the agent should fail"
	or diag $res->as_string;
POST '/delete', { path => 'some/path/some/where' };
ok !$SUCCESS, "attempting to delete a file as the agent should fail"
	or diag $res->as_string;

as_admin;
POST '/download', {}, { path => 'some/path/some/where' };
ok !$SUCCESS, "attempting to create a download as the admin should fail"
	or diag $res->as_string;
POST '/upload', { path => 'some/path/some/where' };
ok !$SUCCESS, "attempting to create an upload as the admin should fail"
	or diag $res->as_string;
POST '/delete', { path => 'some/path/some/where' };
ok !$SUCCESS, "attempting to delete a file as the admin should fail"
	or diag $res->as_string;

as_admin;
GET '/admin/streams';
ok $SUCCESS, "should be able to retrieve streams as admin";
cmp_deeply($RESPONSE, {
	uploads => [],
	downloads => [],
});

as_control;
POST '/upload', { path => 'some/path/some/where' };
ok $SUCCESS, "creating an upload as the controller should succeed"
	or diag $res->as_string;
cmp_deeply($RESPONSE, {
	id      => re(qr/^[0-9a-f]+$/i),
	token   => re(qr/^[0-9a-f]+$/i),
	expires => ignore(),
});

$TOKEN = $RESPONSE->{token};
$id    = $RESPONSE->{id};

as_admin;
GET '/admin/streams';
ok $SUCCESS, "should be able to retrieve streams as admin";
cmp_deeply($RESPONSE, {
	uploads => [{
		id      => $id,
		path    => 'some/path/some/where',
		recv    => 0,
		expires => ignore(),
	}],
	downloads => [],
});

as_agent;
POST "/upload/$id", {
	data => encode_base64("this is the first line\n"),
	eof  => $JSON::true,
};
ok $SUCCESS, "posting data to the upload stream (as the agent) should succeed";
cmp_deeply($RESPONSE, { ok => re(qr/uploaded [0-9]+ bytes \(and finished\)/i) });
ok -d "t/tmp/some/path/some",       "parent directories should be created in file storage";
ok -f "t/tmp/some/path/some/where", "file should be uploaded to file storage successfully";

POST "/upload/$id", {
	data => "this is too much stuff\n",
	eof => $JSON::false,
};
ok !$SUCCESS, "adding more data once sending the EOF should fail";
cmp_deeply($RESPONSE, { error => ignore() });

GET "/download/$id";
ok !$SUCCESS, "attempting to download the app with the same id/token should fail";

as_admin;
GET '/admin/streams';
ok $SUCCESS, "should be able to retrieve streams as admin";
cmp_deeply($RESPONSE, {
	uploads => [],
	downloads => [],
});

as_control;
POST "/download", { path => 'some/path/some/where' };
ok $SUCCESS, "creating a download as the controller should succeed"
	or diag $res->as_string;
cmp_deeply($RESPONSE, {
	id      => ignore(),
	token   => ignore(),
	expires => ignore(),
});
isnt $TOKEN, $RESPONSE->{token}, "should get a new token for the download";
isnt $id,    $RESPONSE->{id},    "should get a new id for the download";

$TOKEN = $RESPONSE->{token};
$id    = $RESPONSE->{id};

as_admin;
GET '/admin/streams';
ok $SUCCESS, "should be able to retrieve streams as admin";
cmp_deeply($RESPONSE, {
	uploads => [],
	downloads => [{
		id      => $id,
		path    => 'some/path/some/where',
		recv    => 0,
		expires => ignore(),
	}],
});

as_agent;
GET "/download/$id";
ok $SUCCESS, "should be able to download the given path";
is $res->decoded_content, "this is the first line\n";

as_control;
ok  -f "t/tmp/some/path/some/where", "file should still be in file storage";
POST "/delete", { path => 'some/path/some/where' };
ok $SUCCESS;
ok !-f "t/tmp/some/path/some/where", "file should not still be in file storage";

#  r.Dispatch("GET /admin/streams", func(r *route.Request) {

done_testing;
