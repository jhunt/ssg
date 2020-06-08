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
my ($id, $CANON, $TOKEN, $AUTH);
sub as_control { $AUTH = 'test-control-token-apqlwoskeij'; }
sub as_admin   { $AUTH = 'test-admin-token-ghtyyfjkrudke'; }
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
  $req->header('Authorization' => "Bearer $AUTH") if  $AUTH;
  $req->header('X-SSG-Token' => $TOKEN)           if !$AUTH && $TOKEN;
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
  $req->header('Authorization' => "Bearer $AUTH") if  $AUTH;
  $req->header('X-SSG-Token' => $TOKEN)           if !$AUTH && $TOKEN;
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
POST '/control', { kind => 'upload', target => 'ssg://cluster1/files' };
ok !$SUCCESS, "attempting to create an upload as the agent should fail"
	or diag $res->as_string;
POST '/control', { kind => 'download', target => 'ssg://cluster1/files/a/file' };
ok !$SUCCESS, "attempting to create a download as the agent should fail"
	or diag $res->as_string;
POST '/control', { kind => 'expunge', target => 'ssg://cluster1/files/a/file' };
ok !$SUCCESS, "attempting to expunge a file as the agent should fail"
	or diag $res->as_string;

#as_admin;
#POST '/control', { kind => 'upload', target => 'ssg://cluster1/files' };
#ok !$SUCCESS, "attempting to create an upload as the admin should fail"
#	or diag $res->as_string;
#POST '/control', { kind => 'download', target => 'ssg://cluster1/files/a/file' };
#ok !$SUCCESS, "attempting to create a download as the admin should fail"
#	or diag $res->as_string;
#POST '/control', { kind => 'expunge', target => 'ssg://cluster1/files/a/file' };
#ok !$SUCCESS, "attempting to expunge a file as the admin should fail"
#	or diag $res->as_string;

as_admin;
GET '/streams';
ok $SUCCESS, "should be able to retrieve streams as admin"
	or diag $res->as_string;
cmp_deeply($RESPONSE, []);

as_control;
POST '/control', { kind => 'upload', target => 'ssg://cluster1/files' };
ok $SUCCESS, "creating an upload as the controller should succeed"
	or diag $res->as_string;
cmp_deeply($RESPONSE, {
	kind    => 'upload',
	id      => re(qr/^[0-9a-v]+$/i),
	token   => re(qr/^[0-9a-v]+$/i),
	canon   => re(qr{^ssg://cluster1/files/.+$}),
	expires => ignore(),
});

$TOKEN = $RESPONSE->{token};
$id    = $RESPONSE->{id};
$CANON = $RESPONSE->{canon};

as_admin;
GET '/streams';
ok $SUCCESS, "should be able to retrieve streams as admin"
	or diag $res->as_string;
cmp_deeply($RESPONSE, [
	{
		kind     => 'upload',
		id       => $id,
		canon    => $CANON,
		received => 0,
		expires  => ignore(),
	},
]);

as_agent;
POST "/blob/$id", {
	data => encode_base64("this is the first line\n"),
	eof  => $JSON::true,
};
ok $SUCCESS, "posting data to the upload stream (as the agent) should succeed"
	or diag $res->as_string;
cmp_deeply($RESPONSE, { ok => re(qr/uploaded [0-9]+ bytes \(and finished\)/i) });

POST "/blob/$id", {
	data => encode_base64("this is too much stuff\n"),
	eof => $JSON::false,
};
ok !$SUCCESS, "adding more data once sending the EOF should fail"
	or diag $res->as_string;
cmp_deeply($RESPONSE, { error => ignore() });

GET "/blob/$id";
ok !$SUCCESS, "attempting to download the blob with the same id/token should fail"
	or diag $res->as_string;

as_admin;
GET '/streams';
ok $SUCCESS, "should be able to retrieve streams as admin"
	or diag $res->as_string;
cmp_deeply($RESPONSE, []);

as_control;
POST "/control", { kind => 'download', target => $CANON };
ok $SUCCESS, "creating a download as the controller should succeed"
	or diag $res->as_string;
cmp_deeply($RESPONSE, {
	kind    => 'download',
	canon   => $CANON,
	id      => ignore(),
	token   => ignore(),
	expires => ignore(),
});
isnt $TOKEN, $RESPONSE->{token}, "should get a new token for the download";
isnt $id,    $RESPONSE->{id},    "should get a new id for the download";

$TOKEN = $RESPONSE->{token};
$id    = $RESPONSE->{id};

as_admin;
GET '/streams';
ok $SUCCESS, "should be able to retrieve streams as admin"
	or diag $res->as_string;
cmp_deeply($RESPONSE, [
	{
		kind     => 'download',
		id       => $id,
		canon    => $CANON,
		received => 0,
		expires  => ignore(),
	},
]);

as_agent;
GET "/blob/$id";
ok $SUCCESS, "should be able to download the blob"
	or diag $res->as_string;
is $res->decoded_content, "this is the first line\n";

as_control;
my $file = $CANON; $file =~ s|^ssg://cluster1/files/||;
ok  -f "t/tmp/$file", "file should still be in file storage";
#POST "/control", { kind => 'expunge', target => $CANON };
#ok $SUCCESS
#	or diag $res->as_string;
#ok !-f "t/tmp/$file", "file should not still be in file storage";

done_testing;
