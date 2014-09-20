$(document).ready(function(){
    $('.overall-table tbody tr').hover(function() {
        var id = $(this).attr("id");
        $(getTeam(id)).addClass('team-hover');
    }, function() {
        var id = $(this).attr("id");
        $(getTeam(id)).removeClass('team-hover');
    });

    $('.overall-table tbody tr').click(function() {
        var selectedClass = 'team-selected';
        var id = $(this).attr("id");

        if(!$(this).hasClass(selectedClass)) {
            $(getTeam(id)).addClass(selectedClass);
        } else {
            $(getTeam(id)).removeClass(selectedClass);
        }
    });

    function getTeam(id) {
        var teamId = id.substring(id.indexOf('-') + 1);
        return ".team-" + teamId;
    }
});
