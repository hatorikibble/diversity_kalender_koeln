#!/usr/bin/env perl

=head1 NAME

csv2ical.pl - converts csv File to ICal data

=head1 SYNOPSIS

./csv2ical.pl > diversity.ics

=cut

use Data::ICal;
use Data::ICal::Entry::Event;
use DateTime;
use DateTime::Event::ICal;
use DateTime::Format::ICal;
use DateTime::Format::Strptime;
use Text::xSV;

my $file     = "../Diversity_Kalender_2017.csv";
my $csv      = Text::xSV->new( sep => ";" );
my $calendar = Data::ICal->new(
    calname    => "Kölner Diversity Kalender 2017",
    auto_uid   => "true",
    rfc_strict => "true"
);
my $strp = DateTime::Format::Strptime->new(
    pattern  => '%d.%m.%Y',
    on_error => 'croak',
);
my $event   = undef;
my $buffer  = undef;
my $summary = undef;

$csv->open_file($file);
$csv->read_header();

while ( $csv->get_row() ) {
    my ( $datum, $feiertag, $religion )
        = $csv->extract(qw(Datum Feiertag Religion));

    my $dt = $strp->parse_datetime($datum);

    #print $dt->strftime('%Y%m%d');

    $event = Data::ICal::Entry::Event->new();
    $event->add_properties(
        summary     => $feiertag,
        description => $religion,
        dtstart     => "DATE:" . $dt->strftime('%Y%m%d'),
        dtend       => "DATE:" . $dt->add( days => 1 )->strftime('%Y%m%d'),
        dtstamp => DateTime::Format::ICal->format_datetime( DateTime->now ),
    );

    $calendar->add_entry($event);

}

$buffer = $calendar->as_string;

# fix for all-day events:
$buffer =~ s/DT(START|END):/DT$1;VALUE=/g;

print $buffer;
