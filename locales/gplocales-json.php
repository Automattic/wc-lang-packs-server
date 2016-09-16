<?php
// +build ignore

require_once( 'gplocales.php' );

echo json_encode( GP_Locales::locales() );
