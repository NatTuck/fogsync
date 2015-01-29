#!/usr/bin/perl
use 5.12.0;
use warnings FATAL => 'all';

use Cwd 'abs_path';
use File::Basename;

my $TEST_SRC = "/usr/share/man";
my $TEST_TMP = "/tmp/fog-test-$$";
my $TEST_EFT = "$TEST_TMP/eft";
my $TEST_DST = "$TEST_TMP/out";

my $FOGT = dirname(abs_path($0)) . "/../bin/fogt";

say $FOGT;

#if (-d $TEST_TMP) {
#    die "Temp dir $TEST_TMP already exists.";
#}

sub runcmd {
    my ($cmd) = @_;
    my $rv = system($cmd);
    if ($rv != 0) {
        say "Exit status:", ($rv >> 8);
        die "runcmd failed";
    }
}

mkdir $TEST_TMP;
chdir $TEST_SRC;

my @files = `find .`;

for my $ff (@files) {
    chomp $ff;
    
    my $cmd = qq{"$FOGT" -d "$TEST_EFT" put "$ff"};
    say $cmd;
    runcmd($cmd);
}

mkdir $TEST_DST;
chdir $TEST_DST;

for my $ff (@files) {
    chomp $ff;

    $ff =~ s/^\./\//;
    $ff =~ s/^\/\//\//;

    my $cmd = qq{"$FOGT" -d "$TEST_EFT" get "$ff"};
    say $cmd;
    runcmd($cmd);
}

say "Directory diff:";
system(qq{diff "$TEST_SRC" "$TEST_DST" | grep -v "^Common subdirectories:"}); 

system("rm -rf $TEST_TMP");
