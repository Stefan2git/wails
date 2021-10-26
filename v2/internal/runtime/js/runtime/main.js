/*
 _       __      _ __    
| |     / /___ _(_) /____
| | /| / / __ `/ / / ___/
| |/ |/ / /_/ / / (__  ) 
|__/|__/\__,_/_/_/____/  
The electron alternative for Go
(c) Lea Anthony 2019-present
*/
/* jshint esversion: 6 */

const Log = require('./log');
const Events = require('./events');
const Init = require('./init');

module.exports = {
	Events: Events,
	ready: Init.ready,
	Log: Log,
};