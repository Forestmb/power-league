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

    // Add the ability to sort the overall standings table
    $('.overall-table table').tablesorter({
        sortList: [[2,1]],
        sortInitialOrder: 'desc',
        sortRestart: true,
        sortStable: true,
        headers: {
            '.overall-header-team': {
                sortInitialOrder: 'asc',
            },
            '.overall-header-league-rank': {
                sortInitialOrder: 'asc',
            }
        }
    });

    var sortedWeeklyTables = 0;
    var numberOfWeeklyTables = $('.weekly table').length;

    // Add the ability to sort the weekly score tables
    $('.weekly table').tablesorter({
        sortList: [[1,1]],
        sortInitialOrder: 'desc',
        sortRestart: true,
        sortStable: true,
        headers: {
            0: { sortInitialOrder: 'asc', }
        }
    // Sort the weekly tables in unison -- a.k.a. if one table is sorted
    // by a column, all weekly tables are sorted by that column
    }).bind("sortEnd", function(sorter) 
    {
        sortedWeeklyTables++;
        var currentSort = sorter.target.config.sortList;
        if (sortedWeeklyTables === 1) {
            $('.weekly table').not(this).trigger("sorton", [ currentSort ]);
        } else if (sortedWeeklyTables === numberOfWeeklyTables) {
            sortedWeeklyTables = 0;
        }
    });

    $(".newsletter-export").click(function(){
        $('.newsletter-container').toggleClass('hidden');
        $('.newsletter-export').toggleClass('selected');
    });
});

// Toggles all-play/power points and stores the last user preference in a cookie
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
