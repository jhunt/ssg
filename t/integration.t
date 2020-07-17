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
use Digest::SHA1;

$ENV{PUBLIC_PORT} = 9077;

my $UA = LWP::UserAgent->new(agent => 'ssg-test/'.strftime("%Y%m%dT%H%M%S", gmtime())."-$$");
my $BASE_URL = "http://127.0.0.1:$ENV{PUBLIC_PORT}";
my ($TOKEN, $AUTH);
sub as_control { $AUTH = 'test-control-token-apqlwoskeij'; }
sub as_monitor { $AUTH = 'test-monitor-token-jjqwhrexck1'; }
sub as_agent   { $AUTH = $TOKEN; }

my ($id, $CANON);
sub local_fs_path {
	my $path = $CANON;
	$path =~ s|^ssg://[^/]*/[^/]*/?|t/tmp/|;
	return $path;
}
sub vault_secret {
	my $path = $CANON;
	$path =~ s|ssg://[^/]*/[^/]*/?|secret/tests/|;
	return $path;
}

sub maybe_json {
	my ($raw) = @_;
	return eval { return decode_json($raw) } or undef;
}

sub atleast {
	my ($n) = @_;
	return code(sub {
		return $_[0] >= $n;
	});
}

my ($req, $res, $SUCCESS, $STATUS, $RESPONSE);
sub GET {
  my ($url) = @_;
  $url =~ s|^/||;

  $req = HTTP::Request->new(GET => "$BASE_URL/$url")
    or die "failed to make [GET /$url] request: $!\n";
  $req->header('Authorization' => "Bearer $AUTH") if $AUTH;
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
  $req->header('Authorization' => "Bearer $AUTH") if $AUTH;
  $req->header('Accept'       => 'application/json');
  $req->header('Content-Type' => 'application/json') if $payload;
  $req->content(encode_json($payload))               if $payload;

  diag $req->as_string if $ENV{DEBUG_HTTP};
  $res = $UA->request($req);
  diag $res->as_string if $ENV{DEBUG_HTTP};
  ($SUCCESS, $STATUS, $RESPONSE) = ($res->is_success, $res->code, maybe_json($res->decoded_content));
}

sub DELETE {
  my ($url) = @_;
  $url =~ s|^/||;

  $req = HTTP::Request->new(DELETE => "$BASE_URL/$url")
    or die "failed to make [DELETE /$url] request: $!\n";
  $req->header('Authorization' => "Bearer $AUTH") if $AUTH;
  $req->header('Accept'       => 'application/json');

  diag $req->as_string if $ENV{DEBUG_HTTP};
  $res = $UA->request($req);
  diag $res->as_string if $ENV{DEBUG_HTTP};
  ($SUCCESS, $STATUS) = ($res->is_success, $res->code);
}

diag "setting up docker-compose integration environment...\n";
system('t/setup');
is $?, 0, 't/setup should exit zero (success)'
  or do { done_testing; exit; };

as_agent;
GET '/buckets';
ok !$SUCCESS, "attempting to list buckets as the agent should fail"
	or diag $res->as_string;

as_monitor;
GET '/buckets';
ok !$SUCCESS, "attempting to list buckets as the monitor should fail"
	or diag $res->as_string;

as_control;
GET '/buckets';
ok $SUCCESS, "attempting to list buckets as the control user should succeed"
	or diag $res->as_string;
cmp_deeply([grep { $_->{key} =~ m/^base-/ } @$RESPONSE], [
	{
		key => 'base-files',
		name => 'Files',
		description => '',
		compression => 'zlib',
		encryption => 'aes256-ctr',
	},
	{
		key => 'base-minio',
		name => 'Minio (S3)',
		description => 'An S3-workalike that puts files in the root of a bucket',
		compression => 'zlib',
		encryption => 'aes256-ctr',
	},
	{
		key => 'base-minio-with-prefix',
		name => 'Minio (S3 /prefix)',
		description => '',
		compression => 'zlib',
		encryption => 'aes256-ctr',
	},
	{
		key => 'base-webdav',
		name => 'WebDAV',
		description => '',
		compression => 'zlib',
		encryption => 'aes256-ctr',
	},
], "/buckets should list only pertinent bucket info, in defined order");

my @buckets = map { $_->{key} } @$RESPONSE;
for my $BUCKET (grep { m/^base-/ } @buckets) {
	last if $ENV{SKIP_PROVIDER_TESTS};
	subtest "$BUCKET bucket" => sub { # {{{
		as_agent;
		POST '/control', { kind => 'upload', target => "ssg://cluster1/$BUCKET" };
		ok !$SUCCESS, "attempting to create an upload as the agent should fail"
			or diag $res->as_string;
		POST '/control', { kind => 'download', target => "ssg://cluster1/$BUCKET/a/file" };
		ok !$SUCCESS, "attempting to create a download as the agent should fail"
			or diag $res->as_string;
		POST '/control', { kind => 'expunge', target => "ssg://cluster1/$BUCKET/a/file" };
		ok !$SUCCESS, "attempting to expunge a file as the agent should fail"
			or diag $res->as_string;
		GET '/streams';
		ok !$SUCCESS, "attempting to retrieve streams as the agent should fail"
			or diag $res->as_string;
		GET '/metrics';
		ok !$SUCCESS, "attempting to retrieve metrics as the agent should fail"
			or diag $res->as_string;
		DELETE '/metrics';
		ok !$SUCCESS, "attempting to clear metrics as the agent should fail"
			or diag $res->as_string;

		as_monitor;
		POST '/control', { kind => 'upload', target => "ssg://cluster1/$BUCKET" };
		ok !$SUCCESS, "attempting to create an upload as the monitor should fail"
			or diag $res->as_string;
		POST '/control', { kind => 'download', target => "ssg://cluster1/$BUCKET/a/file" };
		ok !$SUCCESS, "attempting to create a download as the monitor should fail"
			or diag $res->as_string;
		POST '/control', { kind => 'expunge', target => "ssg://cluster1/$BUCKET/a/file" };
		ok !$SUCCESS, "attempting to expunge a file as the monitor should fail"
			or diag $res->as_string;
		GET '/streams';
		ok !$SUCCESS, "attempting to retrieve streams as the monitor should fail"
			or diag $res->as_string;
		GET '/metrics';
		ok $SUCCESS, "attempting to retrieve metrics as the agent should succeed"
			or diag $res->as_string;
		DELETE '/metrics';
		ok $SUCCESS, "attempting to clear metrics as the agent should succeed"
			or diag $res->as_string;

		as_control;
		GET '/streams';
		ok $SUCCESS, "should be able to retrieve streams as admin"
			or diag $res->as_string;
		cmp_deeply($RESPONSE, [], "no streams should exist");

		as_monitor;
		GET '/metrics';
		ok $SUCCESS, "should be able retrieve metrics as the monitor"
			or diag $res->as_string;
		cmp_deeply($RESPONSE, superhashof({
			$BUCKET => {
				operations => {
					upload => 0,
					download => 0,
					expunge => 0,
				},
				canceled => {
					upload => 0,
					download => 0,
				},
				segments => {
					total => 0,
					bytes => {
						minimum => 0.0,
						maximum => 0.0,
						median => 0.0,
						sigma => 0.0,
					},
				},
				transfer => {
					front => {
						in => 0,
						out => 0,
					},
					back => {
						in => 0,
						out => 0,
					},
				},
			},
		}), "metrics should be initially blank");

		as_control;
		POST '/control', { kind => 'upload', target => "ssg://cluster1/$BUCKET" };
		ok $SUCCESS, "creating an upload as the controller should succeed"
			or diag $res->as_string;
		cmp_deeply($RESPONSE, {
			kind    => 'upload',
			id      => re(qr/^[0-9a-v]+$/i),
			token   => re(qr/^[0-9a-v]+$/i),
			canon   => re(qr{^ssg://cluster1/$BUCKET/.+$}),
			expires => ignore(),
		}, "creating an upload should get us the stream details");

		$TOKEN = $RESPONSE->{token};
		$id    = $RESPONSE->{id};
		$CANON = $RESPONSE->{canon};
		system("./t/vault check '".vault_secret()."'");
		ok $? == 0, 'should have encryption cipher in the vault';

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
		], "our upload stream should be listed");

		as_monitor;
		GET '/metrics';
		ok $SUCCESS, "should be able to retrieve metrics after starting an upload"
			or diag $res->as_string;
		cmp_deeply($RESPONSE, superhashof({
			$BUCKET => {
				operations => {
					upload => 1,
					download => 0,
					expunge => 0,
				},
				canceled => {
					upload => 0,
					download => 0,
				},
				segments => {
					total => 0,
					bytes => {
						minimum => 0.0,
						maximum => 0.0,
						median => 0.0,
						sigma => 0.0,
					},
				},
				transfer => {
					front => {
						in => 0,
						out => 0,
					},
					back => {
						in => 0,
						out => 0,
					},
				},
			},
		}), "metrics should reflect our new upload operation");

		my $DATA = "this is the first line\n";
		as_agent;
		POST "/blob/$id", {
			data => encode_base64($DATA),
			eof  => $JSON::true,
		};
		ok $SUCCESS, "posting data to the upload stream (as the agent) should succeed"
			or diag $res->as_string;
		cmp_deeply($RESPONSE, {
			segments => 1,
			compressed => re(qr/^\d+$/),
			uncompressed => length($DATA),
			sent => length($DATA),
		}, "posting the final segment should return blob details");

		# check the encryption parameters here,
		# so we don't time out the upload...
		for (qw/alg key iv id/) {
			system("./t/vault check '".vault_secret().":$_'");
			ok $? == 0, "should have encryption cipher ($_) in the vault";
		}

		ok  -f local_fs_path(), "file should be in file storage"
			if $BUCKET eq 'base-files';
		is -s local_fs_path(), $RESPONSE->{compressed}, "file should be \$compressed bytes long"
			if $BUCKET eq 'base-files';
		isnt $RESPONSE->{uncompressed}, $RESPONSE->{compressed}, "uncompressed data should be a different size than compressed";

		as_monitor;
		GET '/metrics';
		ok $SUCCESS, "should be able to retrieve metrics after upload a segment"
			or diag $res->as_string;
		cmp_deeply($RESPONSE, superhashof({
			$BUCKET => {
				operations => {
					upload => 1,
					download => 0,
					expunge => 0,
				},
				canceled => {
					upload => 0,
					download => 0,
				},
				segments => {
					total => 1,
					bytes => {
						minimum => length($DATA),
						maximum => length($DATA),
						median => length($DATA),
						sigma => 0.0,
					},
				},
				transfer => {
					front => {
						in => length($DATA),
						out => 0,
					},
					back => {
						in => 0,
						out => atleast(35), # "compressed"
					},
				},
			},
		}), "metrics should reflect our new upload operation");

		POST "/blob/$id", {
			data => encode_base64("this is too much stuff\n"),
			eof => $JSON::false,
		};
		ok !$SUCCESS, "adding more data once sending the EOF should fail"
			or diag $res->as_string;
		cmp_deeply($RESPONSE, { error => ignore() }, "adding more data after EOF should return an error");

		GET "/blob/$id";
		ok !$SUCCESS, "attempting to download the blob with the same id/token should fail"
			or diag $res->as_string;

		as_control;
		GET '/streams';
		ok $SUCCESS, "should be able to retrieve streams as admin"
			or diag $res->as_string;
		cmp_deeply($RESPONSE, [], "no streams should exist after closing our upload");

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
		}, "creating a download should return stream details");
		isnt $TOKEN, $RESPONSE->{token}, "should get a new token for the download";
		isnt $id,    $RESPONSE->{id},    "should get a new id for the download";

		$TOKEN = $RESPONSE->{token};
		$id    = $RESPONSE->{id};

		as_monitor;
		GET '/metrics';
		ok $SUCCESS, "should be able to retrieve metrics after starting a download"
			or diag $res->as_string;
		cmp_deeply($RESPONSE, superhashof({
			$BUCKET => {
				operations => {
					upload => 1,
					download => 1,
					expunge => 0,
				},
				canceled => ignore,
				segments => {
					total => 1,
					bytes => ignore,
				},
				transfer => ignore,
			},
		}), "metrics should reflect our new download operation");

		as_control;
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
		], "our download stream should be listed");

		as_agent;
		GET "/blob/$id";
		ok $SUCCESS, "should be able to download the blob"
			or diag $res->as_string;
		is $res->decoded_content, "this is the first line\n",
			"downloading the blob should return original data posted";

		as_monitor;
		GET '/metrics';
		ok $SUCCESS, "should be able to retrieve metrics after downloading data"
			or diag $res->as_string;
		cmp_deeply($RESPONSE, superhashof({
			$BUCKET => {
				operations => { # SAME
					upload => 1,
					download => 1,
					expunge => 0,
				},
				canceled => { # SAME
					upload => 0,
					download => 0,
				},
				segments => { # SAME
					total => 1,
					bytes => {
						minimum => length($DATA),
						maximum => length($DATA),
						median => length($DATA),
						sigma => 0.0,
					},
				},
				transfer => {
					front => {
						in => length($DATA),
						out => length($DATA),
					},
					back => {
						in => atleast(35), # "compressed"
						out => atleast(35), # "compressed"
					},
				},
			},
		}), "metrics should reflect our new upload operation");

		as_control;
		ok  -f local_fs_path(), "file should still be in file storage"
			if $BUCKET eq 'base-files';
		POST "/control", { kind => 'expunge', target => $CANON };
		ok $SUCCESS, "should be able to expunge the blob"
			or diag $res->as_string;
		ok !-f local_fs_path(), "file should not still be in file storage"
			if $BUCKET eq 'base-files';

		system("./t/vault check '".vault_secret()."'");
		ok $? != 0, 'should no longer have encryption cipher in the vault';

		as_monitor;
		GET '/metrics';
		ok $SUCCESS, "should be able to retrieve metrics after expunging data"
			or diag $res->as_string;
		cmp_deeply($RESPONSE, superhashof({
			$BUCKET => {
				operations => { # SAME
					upload => 1,
					download => 1,
					expunge => 1,
				},
				canceled => { # SAME
					upload => 0,
					download => 0,
				},
				segments => ignore,
				transfer => ignore,
			},
		}), "metrics should reflect our new expunge operation");

		## timeout
		as_control;
		POST '/control', { kind => 'upload', target => "ssg://cluster1/$BUCKET" };
		ok $SUCCESS, "creating an upload as the controller should succeed"
			or diag $res->as_string;
		cmp_deeply($RESPONSE, {
			kind    => 'upload',
			id      => re(qr/^[0-9a-v]+$/i),
			token   => re(qr/^[0-9a-v]+$/i),
			canon   => re(qr{^ssg://cluster1/$BUCKET/.+$}),
			expires => ignore(),
		}, "our upload stream should be listed");

		$TOKEN = $RESPONSE->{token};
		$id    = $RESPONSE->{id};
		$CANON = $RESPONSE->{canon};

		as_monitor;
		GET '/metrics';
		ok $SUCCESS, "should be able to retrieve metrics after starting a second upload"
			or diag $res->as_string;
		cmp_deeply($RESPONSE, superhashof({
			$BUCKET => {
				operations => {
					upload => 2,
					download => 1,
					expunge => 1,
				},
				canceled => {
					upload => 0,
					download => 0,
				},
				segments => ignore,
				transfer => ignore,
			},
		}), "metrics should reflect our new upload operation");

		system("./t/vault check '".vault_secret()."'");
		ok $? == 0, 'should have encryption cipher in the vault';

		as_agent;
		POST "/blob/$id", {
			data => encode_base64("it's---"),
			eof  => $JSON::false,
		};
		ok $SUCCESS, "posting data to the upload stream (as the agent) should succeed"
			or diag $res->as_string;

		as_monitor;
		GET '/metrics';
		ok $SUCCESS, "should be able to retrieve metrics after uploading a segment to the second upload"
			or diag $res->as_string;
		cmp_deeply($RESPONSE, superhashof({
			$BUCKET => {
				operations => { # SAME
					upload => 2,
					download => 1,
					expunge => 1,
				},
				canceled => { # SAME
					upload => 0,
					download => 0,
				},
				segments => {
					total => 2,
					bytes => {
						minimum => 7,
						maximum => length($DATA), # 23
						median => num(15.0, 0.01),
						sigma => num(11.314, 0.01),
					},
				},
				transfer => {
					front => {
						in => length($DATA)+7,
						out => length($DATA),
					},
					back => {
						in => atleast(35), # "compressed"
						out => atleast(35), # "compressed"
						#              ^^
						# out hasn't moved, because we haven't flushed
						# the zlib writer (we wait until we have enough
						# data to make it worthwhile, dictionary-wise).
					},
				},
			},
		}), "metrics should reflect our second segment");

		diag "sleeping 5s waiting for our blob to timeout...";
		sleep 5;
		as_agent
		POST "/blob/$id", {
			data => encode_base64("the rest of the file...\n"),
			eof  => $JSON::true,
		};
		ok !$SUCCESS, "posting data after the stream expires should not succeed"
			or diag $res->as_string;
		system("./t/vault check '".vault_secret()."'");
		ok $? != 0, 'should no longer have encryption cipher in the vault';

		as_monitor;
		GET '/metrics';
		ok $SUCCESS, "should be able to retrieve metrics after uploading a segment to the second upload"
			or diag $res->as_string;
		cmp_deeply($RESPONSE, superhashof({
			$BUCKET => {
				operations => { # SAME
					upload => 2,
					download => 1,
					expunge => 1,
				},
				canceled => { # SAME
					upload => 1,
					download => 0,
				},
				segments => ignore,
				transfer => ignore,
			},
		}), "metrics should reflect our canceled upload");

		# zero-byte file upload test
		as_control;
		POST '/control', { kind => 'upload', target => "ssg://cluster1/$BUCKET" };
		ok $SUCCESS, "starting the zero byte upload should succeed"
			or diag $res->as_string;
		$TOKEN = $RESPONSE->{token};
		$id    = $RESPONSE->{id};
		$CANON = $RESPONSE->{canon};

		as_agent;
		POST "/blob/$id", {
			data => encode_base64(""),
			eof  => $JSON::true,
		};
		ok !$SUCCESS, "posting zero bytes to the upload stream (as the agent) should fail"
			or diag $res->as_string;
	} # }}}
}

sub sha1 {
	my ($file) = @_;
	open my $fh, "<", $file or die "$file: $!\n";
	my $sha = Digest::SHA1->new;
	$sha->addfile($fh);
	close $fh;
	return $sha->hexdigest;
}

sub upload {
	my ($target, $file) = @_;
	open my $fh, "<", $file or die "$file: $!\n";

	as_control;
	POST '/control', { kind => 'upload', target => $target };
	ok $SUCCESS, "starting upload of $file -> $target should succeed"
		or diag $res->as_string;
	$TOKEN = $RESPONSE->{token};
	$id    = $RESPONSE->{id};
	$CANON = $RESPONSE->{canon};

	as_agent;
	while (<$fh>) {
		POST "/blob/$id", {
			data => encode_base64($_),
			eof  => $JSON::false,
		};
		$SUCCESS or last;
	}
	POST "/blob/$id", { eof => $JSON::true };
	ok $SUCCESS, "closing off the upload of $file -> $target should succeed"
		or diag $res->as_string;
}

sub download {
	my ($target) = @_;

	as_control;
	POST '/control', { kind => 'download', target => $target };
	ok $SUCCESS, "starting download of $target should succeed"
		or diag $res->as_string;
	$TOKEN = $RESPONSE->{token};
	$id    = $RESPONSE->{id};
	$CANON = $RESPONSE->{canon};

	as_agent;
	GET "/blob/$id";
	ok $SUCCESS, "download of $target should succeed"
		or diag $res->as_string;

	my $sha = Digest::SHA1->new;
	$sha->add($res->decoded_content);
	return $sha->hexdigest;
}

subtest "fixed-key vault encryption" => sub { # {{{
	my ($a, $b);

	upload 'ssg://cluster1/base-files/a/first/randomized/key/test', 'main.go';
	upload 'ssg://cluster1/base-files/asecond/randomized/key/test', 'main.go';
	$a = sha1('t/tmp/a/first/randomized/key/test');
	$b = sha1('t/tmp/asecond/randomized/key/test');
	isnt $a, $b, 'randomized-key encryption should generate different outputs for identical inputs';

	upload 'ssg://cluster1/fixed-key/a/first/fixed/key/test', 'main.go';
	upload 'ssg://cluster1/fixed-key/asecond/fixed/key/test', 'main.go';
	$a = sha1('t/tmp/a/first/fixed/key/test');
	$b = sha1('t/tmp/asecond/fixed/key/test');
	is $a, $b, 'fixed-key encryption should generate identical outputs for identical inputs';

	upload 'ssg://cluster1/provided-key/a/first/provided/key/test', 'main.go';
	upload 'ssg://cluster1/provided-key/asecond/provided/key/test', 'main.go';
	$a = sha1('t/tmp/a/first/provided/key/test');
	$b = sha1('t/tmp/asecond/provided/key/test');
	is $a, $b, 'fixed-key encryption can take a key/iv, rather than deriving it';
};
# }}}

my @COMPRESS = qw(zlib);
my @ENCRYPT = qw(
	aes128-ctr aes128-cfb aes128-ofb
	aes192-ctr aes192-cfb aes192-ofb
	aes256-ctr aes256-cfb aes256-ofb
);

my $a = sha1('main.go');
for my $c (@COMPRESS) {
	for my $e (@ENCRYPT) {
		upload "ssg://cluster1/x-$c-with-$e/test-$c-with-$e", 'main.go';

		for my $oc (@COMPRESS) {
			for my $oe (@ENCRYPT) {
				my $b = download "ssg://cluster1/x-$oc-with-$oe/test-$c-with-$e";
				is $a, $b, "upload through $c / $e should equal download through $oe / $oc";
			}
		}
	}
}
done_testing;
