function getCookie(name) {
var r = document.cookie.match("\\b" + name + "=([^;]*)\\b");
return r ? r[1] : undefined;
}

$(function(){
  $('.pubButton')
  .click(function(){
    $.ajax({
	 url:'/pub',
	type:'POST',
	data:{'_xsrf':getCookie('_xsrf')},
	success:function(data){location.href="/pub";},
	  error:function(){alert('KO');}
    });
  });	
});
    
    
