#!/usr/bin/env perl

=head1 NAME

csv2ical.pl - converts csv File to ICal data

=head1 SYNOPSIS

./csv2ical.pl > diversity.ics

=cut

use strict;
use warnings;

use Data::ICal;
use Data::ICal::Entry::Event;
use DateTime;
use DateTime::Event::ICal;
use DateTime::Format::ICal;
use DateTime::Format::Strptime;
use Text::xSV;
use Try::Tiny;

my $file = "../Diversity_Kalender_2018.csv";
my $csv = Text::xSV->new( sep => ";" );
my $calendar = Data::ICal->new( calname    => "Kölner Diversity Kalender 2017",
                                auto_uid   => "true",
                                rfc_strict => "true"
);
my $strp =
    DateTime::Format::Strptime->new( pattern => '%d.%m.%Y',
                                     on_error => 'croak', );
my $event   = undef;
my $buffer  = undef;
my $summary = undef;
my $dt_von  = undef;
my $dt_bis  = undef;

$csv->open_file($file);
$csv->read_header();

while ( $csv->get_row() ) {
    my ( $von, $bis, $feiertage, $art, $beschreibung )
        = $csv->extract(qw(von bis Feiertage Art Beschreibung));

    try {
        $dt_von = $strp->parse_datetime($von);
    }
      catch { warn "Problem beim Parsen von $von: $_";
	      
	    };
    if ( defined($bis) ) {
        try {
            $dt_bis = $strp->parse_datetime($bis);
        }
	  catch { warn "Problem beim Parsen von $bis: $_";
		  
		};
    }
    else {
        $dt_bis = $dt_von->add( days => 1 );
    }

    #print $dt->strftime('%Y%m%d');
     $event = Data::ICal::Entry::Event->new();
    $event->add_properties(
            summary     => $feiertage,
            description => sprintf( "(%s) %s", $art, $beschreibung || '' ),
            dtstart     => "DATE:" . $dt_von->strftime('%Y%m%d'),
            dtend       => "DATE:" . $dt_bis->strftime('%Y%m%d'),
            dtstamp => DateTime::Format::ICal->format_datetime( DateTime->now ),
    );

    $calendar->add_entry($event);

} ## end while ( $csv->get_row() )

$buffer = $calendar->as_string;

# fix for all-day events:
$buffer =~ s/DT(START|END):/DT$1;VALUE=/g;
# fix for blank lines
$buffer =~ s/\r\n\s\r\n/\r\n/g;
print $buffer;
