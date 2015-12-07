//= require_tree .
//= require jquery-2.1.4.min.js
//= require highlight.min.js

////////////////// SYNTAX HIGHLIGHTING //////////////////

$(function() {
  hljs.initHighlightingOnLoad();
});

///////////////////// TOGGLE MODAL /////////////////////

function toggleModal() {
  if($('#dl-modal').is(':visible')) {
    $('#dl-modal').fadeOut(200);
  } else {
    $('#dl-modal').fadeIn(200);
  };
}

//////////////////// MODAL BEHAVIOR ////////////////////

$(function() {
  $('a#download, #dl-modal').click(function (e) {
    e.preventDefault();
    toggleModal();
  })
  $(".container").click(function(e) {
      e.stopPropagation();
   });
})

/////////// OPEN / CLOSE RESPONSIVE CONTENTS ///////////

$(function() {
  $('#contents-btn').on('click', function(e) {
    $('#contents').toggleClass('closed');
    $('#contents-btn').toggleClass('open');
  })
})

/////////// ADD/REMOVE CLASS ON CONTENTS BTN ///////////

$(window).on('resize', function() {
  if (document.documentElement.clientWidth > 640) {
    if(!$('#contents').hasClass('closed')) {
      $('#contents').addClass('closed');      
      $('#contents-btn').removeClass('open');
    };
  }
  if (document.documentElement.clientWidth < 640) {
    if($('#contents').hasClass('closed')) {
      $('#contents-btn').removeClass('open');
    };
  }
})

/////////////// HIDE UNUSED NAV SECTIONS ///////////////

$(function(){
  $('ul#contents li:has(ul li.active)').addClass('open');
  $('ul#contents li:has(ul.sub)').addClass('more');
});




$(document).ready(function() {

  /////////////////// CONTENT FADE-IN ///////////////////

  setTimeout(function() {
    $('.fade-in').addClass('show');
  }, 800)
  setTimeout(function() {
    $('.fade-in-fast').addClass('show');
  }, 200)

  ///////////// ADD LINKS TO CONTENT HEADINGS /////////////

  $(".content h2, .content h3, .content h4, .content h5, .content h6").each(function() {
    var link = "<a href=\"#" + $(this).attr("id") + "\"></a>"
    $(this).wrapInner( link );
  })

  //////////////////// SMOOTH SCROLL ////////////////////

  $('a[href^="#"]').on('click',function (e) {
    e.preventDefault();

    var target = this.hash;
    var $target = $(target);

    $('html, body').stop().animate({
      'scrollTop': $target.offset().top
    }, 400, 'swing', function () {
      window.location.hash = target;
    });
  });

});
