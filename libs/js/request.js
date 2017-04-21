/**
 * Created by jino on 2016/5/13.
 */
$(document).ready(function() {

    var loc = window.location
    var uri = 'ws:'
    uri += '//' + loc.host;uri += '/active';
    var ws =new WebSocket(uri);

    ws.onmessage=function(event){

        var msg =JSON.parse(event.data);

        switch (msg.tag){
            case "message":

                alert(msg.body);

                break;

            case "pbjs":

                $(".repolist").select2({
                    data:msg.body,
                    placeholder:{id:'-1',text:"select git repo"},
                    allowClear: true,
                });
                break;

            case "list_branchs":

                $(".branlist").empty();
                $("#srvlist").val("-1").trigger("change");
                $(".branlist").select2({
                    data:msg.body,
                    placeholder:{id:'-1',text:"branch"},
                    allowClear: true,
                });
                break;
            case "list_servers":

                $(".srvlist").empty();
                $(".srvlist").select2({
                    data: msg.body,
                    placeholder: {id: '-1', text: "server"},
                    allowClear: true,
                });
                break;
            case "add_reco":

                item=msg.body;
                $("#queue").append("<tr><td>"+item.pjname+"</td><td>"+item.branch+"</td><td>"+item.branch_test+"</td><td>"+item.server+"</td><td><a class='delrow' id='task"+ item.taskid+"'>x</a></td></tr>");

                break;
            case "mod_status_q":
                item=msg.body;

                $("#task" + item.taskid).text("wait");
                $("#task" + item.taskid).removeClass('delrow');

                $("#status").append("<tr><td>"+ item.pjname+ "</td><td>"+item.branch+"</td><td>"+item.branch_test+"</td><td><samp id=reco"+item.taskid+">"+item.branch_test_status+"</samp></td><td>"+item.owen+"</td></tr>")

                $("#test_queue").append("<tr><td>"+ item.pjname+ "</td><td>"+item.branch+"</td><td>"+item.branch_test+"</td><td>"+item.server+"</td><td>" +
                    "<select id=receive_test"+item.taskid +"><option value="+item.taskid+">wait</option><option value="+item.taskid+">wait->Ok</option></select><button class=but_receive_test>ok</button></td></tr>")

                break;
            case "receive_test":
                item=msg.body
                $("#receive_test"+item.id).empty();
                $("#receive_test"+item.id).append("<option value="+item.id+">"+item.value+"</option>");
                 if (item.value == "wait"){
                     $("#receive_test"+item.id).append("<option value="+item.id+">"+item.value+"->Ok</option>");
                 }
                if (item.value == "testing"){
                    $("#receive_test"+item.id).append("<option value="+item.id+">"+item.value+"->Ok</option>");
                    $("#receive_test"+item.id).append("<option value="+item.id+">"+item.value+"->NotPass</option>");
                }

                $("#reco"+item.id).text(item.value);
                if (item.value == "pass") {

                    $("#task"+item.id).text("pubto");
                    $("#task"+item.id).addClass("requ_upto");

                }else{
                $("#task"+item.id).text(item.value);
                }

                break;

        }

    }



    $(".repolist").select2({placeholder:{id:'-1',text:"select git repo"}});
    $(".branlist").select2({placeholder:{id:'-1',text:"branch"}});
    $(".srvlist").select2({placeholder: {id: '-1', text: "server"}});
    $("#tabs").tabs();


    $("#queue").on('click','.delrow',function(){
        var str= $(this).parent().parent().text().replace(/\s/g,'');
        ws.send(JSON.stringify({tag:"receive_test",body:str}));
        $(this).parent().parent().remove();
    });


    $(".repolist").select2().on("select2:close",function(e){

        ws.send(JSON.stringify({tag:"list_branchs",body:$("#repolist").val()}));
        ws.send(JSON.stringify({tag:"list_servers",body:$("#repolist option:selected").text()}));

    });

    $("#checkreq").click(function(){

        ws.send(JSON.stringify({tag:"add_reco",pjid:$("#repolist").val(),branch:$("#branlist option:selected").text(),srvid:$("#srvlist").val()}));

    });

    $("#apply_test").click(function(){
        ws.send(JSON.stringify({tag:"mod_status",body:"queue"}))
    })


    $(document).on('click','.but_receive_test',function(){
        value= $(this).parent().parent().find("select option:selected").text();
        id= $(this).parent().parent().find("select").val();
        ws.send(JSON.stringify({tag:"receive_test",body:'{"id":'+id+',"value":"'+value+'"}'}));
    })
    

    $(document).on('click','.requ_upto',function(){

        alert($(this).attr("id").replace('task',''));
    })



});
