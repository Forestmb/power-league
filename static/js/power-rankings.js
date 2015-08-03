function getTeamId(id) {
    return id.substring(id.lastIndexOf('-') + 1);
}

function getTeam(id) {
    return ".team-" + getTeamId(id);
}

$(document).ready(function(){
    $('.team-row').hover(function() {
        var id = $(this).attr("id");
        $(getTeam(id)).addClass('team-hover');
    }, function() {
        var id = $(this).attr("id");
        $(getTeam(id)).removeClass('team-hover');
    });

    $('.team-row').click(function() {
        var selectedClass = 'team-selected';
        var id = $(this).attr("id");

        if(!$(this).hasClass(selectedClass)) {
            $(getTeam(id)).addClass(selectedClass);
        } else {
            $(getTeam(id)).removeClass(selectedClass);
        }
    });

    $('.view-scheme').click(function() {
        var schemeId = $(this).attr('data-scheme-id');
        $('.scheme-based').addClass('hidden');
        $('.scheme-' + schemeId).removeClass('hidden');

        $('.scheme-item').removeClass('active');
        $('.scheme-item-' + schemeId).addClass('active');

        document.cookie='PowerPreference=' + schemeId;
    });

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
