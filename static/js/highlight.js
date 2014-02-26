$(document).ready(function(){
    $('.overall-table tbody tr').hover(function() {
        var id = $(this).attr("id");
        $(getTeam(id)).addClass('team-hover');
    }, function() {
        var id = $(this).attr("id");
        $(getTeam(id)).removeClass('team-hover');
    });

    var selectedId = null;
    $('.overall-table tbody tr').click(function() {
        var selectedClass = 'team-selected';
        $('.' + selectedClass).removeClass(selectedClass);
        var id = $(this).attr("id");
        if(selectedId !== id) {
            $(getTeam(id)).addClass(selectedClass);
            selectedId = id;
        } else {
            selectedId = null;
        }
    });

    function getTeam(id) {
        var teamId = id.substring(id.indexOf('-') + 1);
        return ".team-" + teamId;
    }
});
