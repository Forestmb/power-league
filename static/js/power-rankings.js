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

$(document).ready(function(){
    $('.toggle-display').click(function() {
        if($(this).hasClass('showing-power')) {
            $(this).removeClass('showing-power');
            $(this).addClass('showing-record');
            
            document.cookie="PowerPreference=record";
        } else {
            $(this).removeClass('showing-record');
            $(this).addClass('showing-power');

            document.cookie="PowerPreference=power";
        }
        $('.record-visible').toggleClass('hidden');
        $('.power-visible').toggleClass('hidden');
    });

    var preference = document.cookie.replace(/(?:(?:^|.*;\s*)PowerPreference\s*\=\s*([^;]*).*$)|^.*$/, "$1");
    if(preference === 'power') {
        $('.record-visible').addClass('hidden')
        $('.power-visible').removeClass('hidden')
        $('.toggle-display').addClass('showing-power')
        $('.toggle-display').removeClass('showing-record')
    } else if(preference === 'record') {
        $('.power-visible').addClass('hidden')
        $('.record-visible').removeClass('hidden')
        $('.toggle-display').addClass('showing-record')
        $('.toggle-display').removeClass('showing-power')
    }

});
