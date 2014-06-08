#!/usr/bin/perl
use 5.12.0;
use warnings FATAL => 'all';


open my $strace, "<", "strace.txt" or die;
while (<$strace>) {
    chomp;
    next if /futex/;
    /\<(\d+\.\d+)\>$/ or next;

    my $time = 0.0 + $1;

    if ($time > 0.01) {
        say $_;
    }
}
close($strace);
