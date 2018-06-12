( function ( $ ) {
    "use strict";


    // Counter Number
    $('.count').each(function () {
        $(this).prop('Counter',0).animate({
            Counter: $(this).text()
        }, {
            duration: 3000,
            easing: 'swing',
            step: function (now) {
                $(this).text(Math.ceil(now));
            }
        });
    });



    $.ajax({
        url: "/api/stats",
      })
      .done(function( data ) {
           var dailykeys = [];
           var dailyval = [];
           for(var k in data.created_daily) dailykeys.push(k) && dailyval.push(data.created_daily[k]);

           var failedkeys = [];
           var failedval = [];
           for(var k in data.failed_daily) failedkeys.push(k) && failedval.push(data.failed_daily[k]);

           var erroredkeys = [];
           var erroredval = [];
           for(var k in data.errored_daily) erroredkeys.push(k) && erroredval.push(data.errored_daily[k]);

           var succeededkeys = [];
           var succeededval = [];
           for(var k in data.succeeded_daily) succeededkeys.push(k) && succeededval.push(data.succeeded_daily[k]);


           //WidgetChart 1
           var ctx = document.getElementById( "dailychart" );
           ctx.height = 150;
           var myChart = new Chart( ctx, {
               type: 'line',
               data: {
                   labels: dailykeys,
                   type: 'line',
                   datasets: [ {
                       data: dailyval,
                       label: 'Created tasks',
                       backgroundColor: 'transparent',
                       borderColor: 'rgba(255,255,255,.55)',
                   }, ]
               },
               options: {

                   maintainAspectRatio: false,
                   legend: {
                       display: false
                   },
                   responsive: true,
                   tooltips: {
                       mode: 'index',
                       titleFontSize: 12,
                       titleFontColor: '#000',
                       bodyFontColor: '#000',
                       backgroundColor: '#fff',
                       titleFontFamily: 'Montserrat',
                       bodyFontFamily: 'Montserrat',
                       cornerRadius: 3,
                       intersect: false,
                   },
                   scales: {
                       xAxes: [ {
                           gridLines: {
                               color: 'transparent',
                               zeroLineColor: 'transparent'
                           },
                           ticks: {
                               fontSize: 2,
                               fontColor: 'transparent'
                           }
                       } ],
                       yAxes: [ {
                           display:false,
                           ticks: {
                               display: false,
                           }
                       } ]
                   },
                   title: {
                       display: false,
                   },
                   elements: {
                       line: {
                           borderWidth: 1
                       },
                       point: {
                           radius: 4,
                           hitRadius: 10,
                           hoverRadius: 4
                       }
                   }
               }
           } );



      var ctx2 = document.getElementById( "erroredchart" );
      ctx2.height = 150;
      var myChart2 = new Chart( ctx2, {
          type: 'line',
          data: {
              labels: erroredkeys,
              type: 'line',
              datasets: [ {
                  data: erroredval,
                  label: 'Errored tasks',
                  backgroundColor: 'transparent',
                  borderColor: 'rgba(255,255,255,.55)',
              }, ]
          },
          options: {

              maintainAspectRatio: false,
              legend: {
                  display: false
              },
              responsive: true,
              tooltips: {
                  mode: 'index',
                  titleFontSize: 12,
                  titleFontColor: '#000',
                  bodyFontColor: '#000',
                  backgroundColor: '#fff',
                  titleFontFamily: 'Montserrat',
                  bodyFontFamily: 'Montserrat',
                  cornerRadius: 3,
                  intersect: false,
              },
              scales: {
                  xAxes: [ {
                      gridLines: {
                          color: 'transparent',
                          zeroLineColor: 'transparent'
                      },
                      ticks: {
                          fontSize: 2,
                          fontColor: 'transparent'
                      }
                  } ],
                  yAxes: [ {
                      display:false,
                      ticks: {
                          display: false,
                      }
                  } ]
              },
              title: {
                  display: false,
              },
              elements: {
                  line: {
                      borderWidth: 1
                  },
                  point: {
                      radius: 4,
                      hitRadius: 10,
                      hoverRadius: 4
                  }
              }
          }
      } );


 var ctx3 = document.getElementById( "failedchart" );
 ctx3.height = 150;
 var myChart3 = new Chart( ctx3, {
     type: 'line',
     data: {
         labels: failedkeys,
         type: 'line',
         datasets: [ {
             data: failedval,
             label: 'Failed tasks',
             backgroundColor: 'transparent',
             borderColor: 'rgba(255,255,255,.55)',
         }, ]
     },
     options: {

         maintainAspectRatio: false,
         legend: {
             display: false
         },
         responsive: true,
         tooltips: {
             mode: 'index',
             titleFontSize: 12,
             titleFontColor: '#000',
             bodyFontColor: '#000',
             backgroundColor: '#fff',
             titleFontFamily: 'Montserrat',
             bodyFontFamily: 'Montserrat',
             cornerRadius: 3,
             intersect: false,
         },
         scales: {
             xAxes: [ {
                 gridLines: {
                     color: 'transparent',
                     zeroLineColor: 'transparent'
                 },
                 ticks: {
                     fontSize: 2,
                     fontColor: 'transparent'
                 }
             } ],
             yAxes: [ {
                 display:false,
                 ticks: {
                     display: false,
                 }
             } ]
         },
         title: {
             display: false,
         },
         elements: {
             line: {
                 borderWidth: 1
             },
             point: {
                 radius: 4,
                 hitRadius: 10,
                 hoverRadius: 4
             }
         }
     }
 } );



  var ctx4 = document.getElementById( "succeededchart" );
  ctx4.height = 150;
  var myChart4 = new Chart( ctx4, {
      type: 'line',
      data: {
          labels: succeededkeys,
          type: 'line',
          datasets: [ {
              data: succeededval,
              label: 'Succeeded tasks',
              backgroundColor: 'transparent',
              borderColor: 'rgba(255,255,255,.55)',
          }, ]
      },
      options: {

          maintainAspectRatio: false,
          legend: {
              display: false
          },
          responsive: true,
          tooltips: {
              mode: 'index',
              titleFontSize: 12,
              titleFontColor: '#000',
              bodyFontColor: '#000',
              backgroundColor: '#fff',
              titleFontFamily: 'Montserrat',
              bodyFontFamily: 'Montserrat',
              cornerRadius: 3,
              intersect: false,
          },
          scales: {
              xAxes: [ {
                  gridLines: {
                      color: 'transparent',
                      zeroLineColor: 'transparent'
                  },
                  ticks: {
                      fontSize: 2,
                      fontColor: 'transparent'
                  }
              } ],
              yAxes: [ {
                  display:false,
                  ticks: {
                      display: false,
                  }
              } ]
          },
          title: {
              display: false,
          },
          elements: {
              line: {
                  borderWidth: 1
              },
              point: {
                  radius: 4,
                  hitRadius: 10,
                  hoverRadius: 4
              }
          }
      }
  } );

});




} )( jQuery );
