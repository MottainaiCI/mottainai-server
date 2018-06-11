var longDayNames = [
  "Sunday",
  "Monday",
  "Tuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday"
];

var shortDayNames = [ 
  "Sun",
  "Mon",
  "Tue",
  "Wed",
  "Thu",
  "Fri",
  "Sat"
];

var shortMonthNames = [
  "---",
  "Jan",
  "Feb",
  "Mar",
  "Apr",
  "May",
  "Jun",
  "Jul",
  "Aug",
  "Sep",
  "Oct",
  "Nov",
  "Dec"
];

var longMonthNames = [
  "---",
  "January",
  "February",
  "March",
  "April",
  "May",
  "June",
  "July",
  "August",
  "September",
  "October",
  "November",
  "December"
];

/* All of the different supported parts of a time format string, with some user-friendly metadata. There's a lot that could be done better here. */
var fmtParts = {
    "Year": [
      {
        "short-name": "Year",
        "name": "Four-digit year",
        "fmt": "2006",
      },
      {
        "short-name": "Year",
        "name": "Two-digit year",
        "fmt": "06",
      }
    ],
    "Month": [
      {
        "short-name": "Month",
        "name": "The full name of the month",
        "fmt": "January",
      },
      {
        "short-name": "Month",
        "name": "The three-letter abbreviation for the month",
        "fmt": "Jan",
      },
      {
        "short-name": "Month",
        "name": "The number of the month, with no leading 0",
        "fmt": "1",
      },
      {
        "short-name": "Month",
        "name": "The number of the month, padded with a leading 0",
        "fmt": "01",
      }
    ],
    "Day": [
      {
        "short-name": "Day",
        "name": "The number of the day, with no leading 0",
        "fmt": "2",
      },
      {
        "short-name": "Day",
        "name": "The number of the day, padded with a leading space",
        "fmt": "_2",
      },
      {
        "short-name": "Day",
        "name": "The number of the day, padded with a leading zero",
        "fmt": "02",
      },
      {
        "short-name": "Day",
        "name": "The three-letter abbreviation for the day of the week",
        "fmt": "Mon",
      },
      {
        "short-name": "Day",
        "name": "The full name of the day of the week",
        "fmt": "Monday",
      },
    ],
    "Hour": [
      {
        "short-name": "Hour",
        "name": "The twenty-four hour time",
        "fmt": "15",
      },
      {
        "short-name": "Hour",
        "name": "The twelve hour time, with no leading zero",
        "fmt": "3",
      },
      {
        "short-name": "Hour",
        "name": "The twelve hour time, padded with a leading zero",
        "fmt": "03",
      },
    ],
    "Minute": [
      {
        "short-name": "Minutes",
        "name": "The number of minutes, with no leading zero",
        "fmt": "4",
      },
      {
        "short-name": "Minutes",
        "name": "The number of minutes, padded with a leading zero",
        "fmt": "04",
      },
    ],
    "Seconds": [
      {
        "short-name": "Seconds",
        "name": "The number of seconds, with no leading zero",
        "fmt": "5",
      },
      {
        "short-name": "Seconds",
        "name": "The number of seconds, padded with a leading zero",
        "fmt": "05",
      },
    ],
    "AM/PM": [
      {
        "short-name": "AM/PM",
        "name": "The time of day (AM or PM), in all upper-case",
        "fmt": "PM",
      },
      {
        "short-name": "AM/PM",
        "name": "The time of day (am or pm), in all lower-case",
        "fmt": "pm",
      },
    ],
    "Milliseconds": [
      {
        "short-name": "Milliseconds",
        "name": "The number of milliseconds - must be exactly three digits long",
        "fmt": ".000",
      },
      {
        "short-name": "Milliseconds",
        "name": "The number of milliseconds - may be up to three digits long",
        "fmt": ".999",
      }
    ],
    "Microseconds": [
      {
        "short-name": "Microseconds",
        "name": "The number of microseconds - must be exactly six digits long",
        "fmt": ".000000",
      },
      {
        "short-name": "Microseconds",
        "name": "The number of microseconds - may be up to six digits long",
        "fmt": ".999999",
      }
    ],
    "Timezone": [
      {
        "short-name": "Timezone",
        "name": "The name of the time zone",
        "fmt": "MST",
      },   
      {
        "short-name": "Timezone",
        "name": "ISO 8601 - either the numeric offset or a 'Z' for UTC - with minutes",
        "fmt": "Z0700",
      },
      {
        "short-name": "Timezone",
        "name": "ISO 8601 - either the numeric offset or a 'Z' for UTC - with seconds",
        "fmt": "Z070000",
      },
      {
        "short-name": "Timezone",
        "name": "ISO 8601 - either the numeric offset or a 'Z' for UTC - with a colon separator",
        "fmt": "Z07:00",
      },
      {
        "short-name": "Timezone",
        "name": "ISO 8601 - either the numeric offset or a 'Z' for UTC - with seconds and colon separators",
        "fmt": "Z07:00:00",
      },
      {
        "short-name": "Timezone",
        "name": "The offset from UTC in hours and minutes",
        "fmt": "-0700",
      },
      {
        "short-name": "Timezone",
        "name": "The offset from UTC in hours, minutes and seconds",
        "fmt": "-070000",
      },
      {
        "short-name": "Timezone",
        "name": "The offset from UTC in hours and minutes, with a colon delimiter",
        "fmt": "-07:00",
      },
      {
        "short-name": "Timezone",
        "name": "The offset from UTC in hours, minutes and seconds, with a colon delimiter",
        "fmt": "-07:00:00",
      },
      {
        "short-name": "Timezone",
        "name": "The offset from UTC in hours",
        "fmt": "-07",
      },
      {
        "short-name": "Timezone",
        "name": "ISO 8601 - either the numeric offset or a 'Z' for UTC - only hours",
        "fmt": "Z07",
      },
    ]
  };

  /* Split a time format string into the different parts. If characters don't match any formats, they're assumed to be literals. 

     This is basically a port of time.nextStdChunk (https://golang.org/src/time/format.go#L130), except it splits the entire format string into a list of parts with metadata. The definitions of every part can be found in fmtParts, organized by type. */
  function tokenizeFormat(str) {
    var tokens = [];
    var i = 0;
    while (i < str.length) {
      switch (str[i]) {
        case "J":
          if (str.slice(i, i+7) == "January") {
            tokens.push(fmtParts["Month"][0])
            i += 7
          } else if (str.slice(i, i+3) == "Jan") {
            tokens.push(fmtParts["Month"][1])
            i += 3
          } else {
            tokens.push({"name":"Unknown", "fmt":str[i]})
            i += 1
          }
          break;
        case "M":
          if (str.slice(i, i+6) == "Monday") {
            tokens.push(fmtParts["Day"][4])
            i += 6
          } else if (str.slice(i, i+3) == "Mon") {
            tokens.push(fmtParts["Day"][3])
            i += 3
          } else if (str.slice(i, i+3) == "MST") {
            tokens.push(fmtParts["Timezone"][0])
            i += 3
          } else {
            tokens.push({"name":"Unknown", "fmt":str[i]})
            i += 1
          }
          break;
        case "0":
          switch (str[i+1]) {
            case "1":
              tokens.push(fmtParts["Month"][3])
              i += 2
              break;
            case "2":
              tokens.push(fmtParts["Day"][2])
              i += 2
              break;
            case "3":
              tokens.push(fmtParts["Hour"][2])
              i += 2
              break;
            case "4":
              tokens.push(fmtParts["Minute"][1])
              i += 2
              break;
            case "5":
              tokens.push(fmtParts["Seconds"][1])
              i += 2
              break;
            case "6":
              tokens.push(fmtParts["Year"][1])
              i += 2
              break;
            default:
              tokens.push({"name":"Unknown", "fmt":str[i]})
              i += 1
              break;
          } 
          break;
        case "1":
          if (str.slice(i, i+2) == "15") {
            tokens.push(fmtParts["Hour"][0])
            i += 2
          } else {
            tokens.push(fmtParts["Month"][2])
            i += 1
          }
          break;
        case "2":
          if (str.slice(i, i+4) == "2006") {
            tokens.push(fmtParts["Year"][0])
            i += 4
          } else {
            tokens.push(fmtParts["Day"][0]) 
            i += 1
          }
          break;
        case "_":
          if (str[i+1] == "2") {
            tokens.push(fmtParts["Day"][1])
            i += 2
          } else {
            tokens.push({"name":"Unknown", "fmt":str[i]})
            i += 1
          }
          break;
        case "3":
          tokens.push(fmtParts["Hour"][1])
          i += 1
          break;
        case "4":
          tokens.push(fmtParts["Minute"][0])
          i += 1
          break;
        case "5":
          tokens.push(fmtParts["Seconds"][0])
          i += 1
          break;
        case "P":
          if (str[i+1] == "M") {
            tokens.push(fmtParts["AM/PM"][0])
            i += 2
          } else {
            tokens.push({"name":"Unknown", "fmt":str[i]})
            i += 1
          }
          break;
        case "p":
          if (str[i+1] == "m") {
            tokens.push(fmtParts["AM/PM"][1])
            i += 2
          } else {
            tokens.push({"name":"Unknown", "fmt":str[i]})
            i += 1
          }
          break;
        case "-":
          if (str.slice(i, i+7) == "-070000") {
            tokens.push(fmtParts["Timezone"][6])
            i += 7
          } else if (str.slice(i, i+9) == "-07:00:00") {
            tokens.push(fmtParts["Timezone"][8])
            i += 9
          } else if (str.slice(i, i+5) == "-0700") {
            tokens.push(fmtParts["Timezone"][5])
            i += 5
          } else if (str.slice(i, i+6) == "-07:00") {
            tokens.push(fmtParts["Timezone"][7])
            i += 6
          } else if (str.slice(i, i+3) == "-07") {
            tokens.push(fmtParts["Timezone"][9])
            i += 3
          } else {
            tokens.push({"name":"Unknown", "fmt":str[i]})
            i += 1
          }
          break;
        case "Z":
          if (str.slice(i, i+7) == "Z070000") {
            tokens.push(fmtParts["Timezone"][2])
            i += 7
          } else if (str.slice(i, i+9) == "Z07:00:00") {
            tokens.push(fmtParts["Timezone"][4])
            i += 9
          } else if (str.slice(i, i+5) == "Z0700") {
            tokens.push(fmtParts["Timezone"][1])
            i += 5
          } else if (str.slice(i, i+6) == "Z07:00") {
            tokens.push(fmtParts["Timezone"][3])
            i += 6
          } else {
            tokens.push({"name":"Unknown", "fmt":str[i]})
            i += 1
          }
          break;
        case ".":
          if (str[i+1] == "9") {
            var j = i+1;
            for (; j < str.length; j++) {
              if (str[j] != "9") { break }   
            }
            if (str[j] != " " && !isNaN(str[j]) && j < str.length) {
              tokens.push({"name":"Unknown", "fmt":str[i]}); 
              i += 1; 
              break;
            }
            tokens.push({"short-name": "Fractional seconds", "fmt":str.slice(i, j)})
            i = j
          } else if (str[i+1] == "0") {
            var j = i+1;
            for (; j < str.length; j++) {
              if (str[j] != "0") { break }   
            }
            if (str[j] != " " && !isNaN(str[j]) && j < str.length) {
              tokens.push({"name":"Unknown", "fmt":str[i]});
              i += 1; 
              break;
            }
            tokens.push({"short-name": "Fractional seconds", "fmt":str.slice(i, j)})
            i = j
          } else {
            tokens.push({"name":"Unknown", "fmt":str[i]})
            i += 1
          }
          break;
        default:
          tokens.push({"name":"Unknown", "fmt":str[i]}) 
          i += 1
          break;
      }
    }
    return tokens
  }

  function getNum(str, pos, fixed) {
    var val;
    var len;
    if (fixed) {
      val = parseInt(str.slice(pos, pos+2));
      len = 2;
    } else if (str[pos+1] >= '0' && str[pos+1] <= '9') {
      val = parseInt(str.slice(pos, pos+2));
      len = 2;
    } else {
      val = parseInt(str[pos]);
      len = 1;
    }
    return {
      "val": val,
      "len": len,
      "failed": (val == NaN)
    }
  }

  function getSpaceNum(str, pos) {
    var val;
    var spaceOffset = 0;
    if (str[pos] == " ") {
      spaceOffset = 1;
    } 
    ret = getNum(str, pos + spaceOffset);
    ret.len += spaceOffset;
    return ret;
  }

  /* Find the index of the element of `arr` which occurs in `str` at position `pos`. */
  function lookup(arr, str, pos) {
    for (var i=0; i < arr.length; i++) {
      if (arr[i].length + pos <= str.length) {
        if (str.slice(pos, pos+arr[i].length).toLowerCase() == arr[i].toLowerCase()) {
          return {
            val: i,
            len: arr[i].length,
            failed: false
          }
        }
      }
    }
    return { failed: true } 
  }

  /* Parse a fractional second, beginning with a dot `.`
     If `minNums` > 0, there must be exactly `minNums` after the 
     dot. Otherwise all numeric characters will be consumed. */
  function getNanos(str, pos, minNums) {
    if (str[pos] != ".") {
      if (minNums == 0) {
        return {"val": 0, "len":0};
      } else {
        return {"val": 0, "len":0, "failed": true};
      }
    }
    var i;
    for (i=pos+1; i<str.length; i++){ 
      if (str[i] < '0' || str[i] > '9') {
        break;
      }
    }
    if (minNums != 0 && i-pos != minNums) {
      return { "failed": true, "len": 0, "val": 0 };
    }
    var val = parseInt(str.slice(pos+1, i));
    val = val * Math.pow(10, 8-(i-(pos+1)));
    return {"failed": false, "len": i-pos, "val": val};
  }

  /* Given a list of format parts from tokenizeFormat, parse the given timeStr into it's constituent parts. This is basically a port of time.Parse (https://golang.org/src/time/format.go#L732).

     Note: We don't have the list of timezone offsets, so this method doesn't mess with the hour and minute fields. The timezone name (or offset if it's an offset from UTC) will be parsed out and returned, but the hour and minute fields will *not* be adjusted appropriately. 
  */
  function parseTimeString(timeStr, tokens) {
    var year = 0;
    var month = 1;
    var day = 1;
    var dayName = "";
    var hour = 0; 
    var min = 0;
    var sec = 0;
    var nsec = 0;
    var zoneOffset = -1;
    var zoneName = "";
    var pos = 0;
    var err = "";
    var failed = false;
    var amSet = false;
    var pmSet = false;
    for (var i=0; i<tokens.length; i++) {
      var token = tokens[i];
      if (failed) {
        break;
      }
      switch (token["short-name"]) {
        case "Year":
            if ((timeStr.length - pos) < token.fmt.length) {
              err = "Not enough characters for year";
              failed = true;
              break; 
            } 
            if (token.fmt.length == 2) {
              var yearParse = parseInt(timeStr.slice(pos, pos+2));
              if (yearParse == NaN) {
                err = "Invalid number for year";
                failed = true;
                break;
              }
              if (yearParse >= 69) {
                year = yearParse + 1900;
              } else {
                year = yearParse + 2000;
              }
              pos += 2;
            } else {
              year = parseInt(timeStr.slice(pos, pos+4));
              if (year == NaN) {
                err = "Invalid number for year";
                failed = true;
                break;
              }
              pos += 4;
            } 
          break;
        case "Month":
          if ((timeStr.length - pos) < token.fmt.length) {
            err = "Not enough characters for month";
            failed = true;
            break;
          }
          var monthVal;
          if (token["fmt"] == "01") {
            monthVal = getNum(timeStr, pos, true);
          } else if (token["fmt"] == "1") {
            monthVal = getNum(timeStr, pos, false);
          } else if (token["fmt"] == "January") {
            monthVal = lookup(longMonthNames, timeStr, pos);
          } else if (token["fmt"] == "Jan") {
            monthVal = lookup(shortMonthNames, timeStr, pos);
          }
          if (monthVal.failed) {
            err = "Invalid value for month";
            failed = true;
            break;
          }
          pos += monthVal.len;
          month = monthVal.val;
          break;
        case "Day":
          if ((timeStr.length - pos) < token.fmt.length) {
            err = "Not enough characters for day";
            failed = true;
            break;
          }
          var dayVal;
          if (token["fmt"] == "2") {
            dayVal = getNum(timeStr, pos, false);
            day = dayVal.val;
          } else if (token["fmt"] == "02") {
            dayVal = getNum(timeStr, pos, true);
            day = dayVal.val;
          } else if (token["fmt"] == "_2") {
            dayVal = getSpaceNum(timeStr, pos); 
            day = dayVal.val;
          } else if (token["fmt"] == "Mon") {
            // Advance the position but don't do anything
            dayVal = lookup(shortDayNames, timeStr, pos);
            dayName = shortDayNames[dayVal.val];
          } else if (token["fmt"] == "Monday") {
            // Advance the position but don't do anything
            dayVal = lookup(longDayNames, timeStr, pos);
            dayName = shortDayNames[dayVal.val];
          }
          if (dayVal.failed) {
            err = "Invalid value for day";
            failed = true;
            break;
          }
          pos += dayVal.len;
          break;
        case "Hour":
          if ((timeStr.length - pos) < token.fmt.length) {
            err = "Not enough characters for hour";
            failed = true;
            break;
          }
          var hourVal = getNum(timeStr, pos, (token["fmt"] == "03"));
          if (hourVal.failed) {
            err = "Invalid value for hours";
            failed = true;
            break;
          }
          pos += hourVal.len;
          hour = hourVal.val;
          break;
        case "Minutes":
          if ((timeStr.length - pos) < token.fmt.length) {
            err = "Not enough characters for minute";
            failed = true;
            break;
          }
          var minVal = getNum(timeStr, pos, (token["fmt"] == "04"));
          if (minVal.failed) {
            err = "Invalid value for minutes";
            failed = true;
            break;
          }
          pos += minVal.len;
          min = minVal.val;
          break;
        case "Seconds":
          if ((timeStr.length - pos) < token.fmt.length) {
            err = "Not enough characters for seconds";
            failed = true;
            break;
          }
          var secVal = getNum(timeStr, pos, (token["fmt"] == "05"));
          if (secVal.failed) {
            err = "Invalid value for seconds";
            failed = true;
            break;
          }
          pos += secVal.len;
          sec = secVal.val;

          // Handle fractional seconds that aren't in the pattern
          if (i == tokens.length-1 || tokens[i+1].fmt[0] != ".") {
            if (pos < timeStr.length && timeStr[pos] == ".") {
              nanoVal = getNanos(timeStr, pos, 0);
              if (nanoVal.failed) {
                err = "Invalid fractional seconds";
                failed = true; 
                break;
              }
              nsec = nanoVal.val;
              pos += nanoVal.len;
            } 
          }
          break;
        case "AM/PM":
          if ((timeStr.length - pos) < 2) {
            err = "Not enough characters for AM/PM";
            failed = true;
            break;
          }
          if (timeStr.slice(pos, pos+2) == "AM" && token.fmt == "PM") {
            amSet = true;
          } else if (timeStr.slice(pos, pos+2) == "PM" && token.fmt == "PM") {
            pmSet = true;
          } else if (timeStr.slice(pos, pos+2) == "am" && token.fmt == "pm") {
            amSet = true;
          } else if (timeStr.slice(pos, pos+2) == "pm" && token.fmt == "pm") {
            pmSet = true;
          } else {
            err = "Invalid value for AM/PM";
            failed = true;
            break;
          }
          pos += 2;
          break;
        case "Fractional seconds":
          var nanoVal;
          if (token["fmt"][1] == "0") {
            nanoVal = getNanos(timeStr, pos, token["fmt"].length);
          } else {
            nanoVal = getNanos(timeStr, pos, 0);
          }
          if (nanoVal.failed) {
            err = "Invalid fractional seconds";
            failed = true; 
            break;
          }
          nsec = nanoVal.val;
          pos += nanoVal.len;
          break;
        case "Timezone":
          if (token["fmt"] == "MST") {
            var tzSlice = timeStr.slice(pos, pos+3);
            if (timeStr.length - pos >= 3 && tzSlice == "UTC") {
              zoneOffset = 0;
              zoneName = "UTC";
              pos += 3;
              break;
            }
            if (timeStr.length - pos >= 4 && (timeStr.slice(pos, pos+4) == "ChST" || timeStr.slice(pos, pos+4) == "MeST")) {
              zoneOffset = 0;
              zoneName = timeStr.slice(pos, pos+4);
              pos += 4;
              break;
            }
            if (timeStr.length - pos > 4 && tzSlice == "GMT") {
              var sign = timeStr.slice(pos+3, pos+4);
              if (sign == "+" || sign == "-") {
                var j = pos+4;
                for (; j < timeStr.length; j++) {
                  var nextChar = timeStr.slice(j, j+1);
                  if (isNaN(nextChar) || nextChar == " ") {
                    break; 
                  }
                }
                var offset = parseInt(timeStr.slice(pos+4, j));
                if (sign == "-") {
                  offset = offset * -1;
                }
                if (offset == 0 || offset < -14 || offset > 12) {
                  zoneOffset = 0;
                  zoneName = "GMT";
                  pos += 3;
                  break;
                }
                zoneOffset = offset * 60 * 60;
                zoneName = timeStr.slice(pos, j);
                pos = j;
                break;
              } else {
                zoneOffset = 0;
                zoneName = "GMT";
                pos += 3;
                break;
              }
            }
            var j = pos;
            for (; j < timeStr.length; j++) {
              var nextChar = timeStr.slice(j, j+1);
              if (/[^A-Z]/.test(nextChar)) {
                break;
              }
            }
            if (j-pos == 3) {
              zoneName = timeStr.slice(pos, pos+3);
              pos += 3;
              break;
            } else if ((j-pos == 4 || j-pos == 5) && timeStr.slice(j, j+1) == "T") {
              zoneName = timeStr.slice(pos, j);
              pos = j;
              break;
            }
            err = "Invalid time zone";
            failed = true;
            break;
          } else {
            if ((token["fmt"] == "Z0700" || token["fmt"] == "Z07" || token["fmt"] == "Z07:00" ) && timeStr.slice(pos, pos+1) == "Z") {
              zoneOffset = 0;
              zoneName = "UTC";
              pos += 3;
              break;
            }
            var tzSign, tzHour, tzMin, tzSeconds
            if (token["fmt"] == "Z07:00" || token["fmt"] == "-07:00") {
              if (timeStr.length - pos < 6) {
                err = "Not enough characters for timezone - expected 6";
                failed = true;
                break;
              }
              if (timeStr.slice(pos+3, pos+4) != ":") {
                err = "Expected colon delimiter for timezone";
                failed = true;
                break;
              }
              tzSign = timeStr.slice(pos, pos+1);
              tzHour = timeStr.slice(pos+1, pos+3); 
              tzMin = timeStr.slice(pos+4, pos+6); 
              tzSeconds = "00";
              pos += 6; 
            } else if (token["fmt"] == "-07" || token["fmt"] == "Z07") {
              if (timeStr.length - pos < 3) {
                err = "Not enough characters for timezone - expected 3";
                failed = true;
                break;
              }
              tzSign = timeStr.slice(pos, pos+1);
              tzHour = timeStr.slice(pos+1, pos+3); 
              tzMin = "00"; 
              tzSeconds = "00";
              pos += 3;
            } else if (token["fmt"] == "Z07:00:00" || token["fmt"] == "-07:00:00") {
              if (timeStr.length - pos < 9) {
                err = "Not enough characters for timezone - expected 9";
                failed = true;
                break;
              }
              if (timeStr.slice(pos+3, pos+4) != ":" || timeStr.slice(pos+6, pos+7) != ":") {
                err = "Expected colon delimiters for timezone";
                failed = true;
                break;
              }
              tzSign = timeStr.slice(pos, pos+1);
              tzHour = timeStr.slice(pos+1, pos+3); 
              tzMin = timeStr.slice(pos+4, pos+6); 
              tzSeconds = timeStr.slice(pos+7, pos+9);
              pos += 9;
            } else if (token["fmt"] == "Z070000" || token["fmt"] == "-070000") {
              if (timeStr.length - pos < 7) {
                err = "Not enough characters for timezone - expected 7";
                failed = true;
                break;
              }
              tzSign = timeStr.slice(pos, pos+1);
              tzHour = timeStr.slice(pos+1, pos+3); 
              tzMin = timeStr.slice(pos+3, pos+5); 
              tzSeconds = timeStr.slice(pos+5, pos+7);
              pos += 9;
            } else {
              if (timeStr.length - pos < 5) {
                err = "Not enough characters for timezone - expected 5";
                failed = true;
              }
              tzSign = timeStr.slice(pos, pos+1);
              tzHour = timeStr.slice(pos+1, pos+3); 
              tzMin = timeStr.slice(pos+3, pos+5); 
              tzSeconds = "00";
              pos += 5;
            }
            hr = parseInt(tzHour);
            if (hr == NaN) {
              err = "Unable to parse timezone hour";
              failed = true;
              break;
            }
            mm = parseInt(tzMin);
            if (mm == NaN) {
              err = "Unable to parse timezone minutes";
              failed = true;
              break;
            }
            ss = parseInt(tzSeconds);
            if (ss == NaN) {
              err = "Unable to parse timezone seconds";
              failed = true;
              break;
            }
            zoneOffset = (hr*60 + mm)*60 + ss;
            if (tzSign == "-") {
              zoneOffset = -1 * zoneOffset;
            } else if (tzSign != "+") {
             err = "Invalid sign for timezone";
             failed = true;
             break;
            }
          } 
          break;
        default:
          for (j=0; j < token["fmt"].length; j++) {
            if (timeStr.slice(pos, pos+1) != token["fmt"].slice(j, j+1)) {
              err = "String ended unexpectedly";
              failed = true;
              break;
            }
            if (token["fmt"].slice(j, j+1) == " ") {
              for (var k=pos; k < timeStr.length; k++) {
                if (timeStr.slice(k, k+1) == " ") {
                  pos += 1;
                  if (tokens[i+1]["fmt"] == "_2" && timeStr.slice(k+1, k+2) != " " && timeStr.slice(k+1, k+2) != "0") {
                    break;
                  }  
                } else {
                  break;
                }
              }
            } else {
              pos += 1;
            } 
          }
      } 
    }

    if (pmSet && hour < 12) {
      hour += 12
    } else if (amSet && hour == 12) {
      hour = 0
    }
 
    return {
      "Year": year,
      "Month": month,
      "WeekDay": dayName,
      "Day": day,
      "Hour": hour,
      "Minutes": min,
      "Seconds": sec,
      "Nanos": nsec,
      "TZSecs": zoneOffset,
      "TZName": zoneName,
      "err": err,
      "failed": failed
    }
  }

/* We want to import this file as a module for Mocha testing, but also include it client-side for the website itself */
if (typeof module != 'undefined') {
  module.exports = {"parseTimeString": parseTimeString, "tokenizeFormat": tokenizeFormat}
}
