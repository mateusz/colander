<?php

$factor = !empty($_GET['f']) ? (int)$_GET['f'] : 1;

for ($j = 0; $j<$factor; $j++) {
	for ($i = 0; $i<10000000; $i++) {
		sin($i);
	}
}
