var nameOrder = true; 
var gradeOrder = true; 
// Colors any really bad or really good ingredients accordingly
window.onload = colorRows(); 

// This next section takes care of organizing the table by name or grade

document.getElementById("name").onclick = function() {
    var rowGrades = document.getElementsByClassName("grade"); 
    var rowNames = document.getElementsByClassName("name");
    
    var list = []; 
    var nameList = []; 
    for (var i = 0; i < rowGrades.length; i++) {
        var n = rowNames[i].innerHTML; 
        var g = rowGrades[i].innerHTML; 
        tempObj = {name: n, grade: g}; 
        list.push(tempObj); 
        nameList.push(n); 
    }
    
    var templist = orderByName(list, nameList, nameOrder); 
    nameOrder = !nameOrder;  
    
    for (var i = 0; i < templist.length; i++) {
        rowNames[i].innerHTML = templist[i].name; 
        rowGrades[i].innerHTML = templist[i].grade; 
    }
    colorRows(); 
    
}

document.getElementById("grade").onclick = function() {
    var rowGrades = document.getElementsByClassName("grade"); 
    var rowNames = document.getElementsByClassName("name");
    
    var list = []; 
    for (var i = 0; i < rowGrades.length; i++) {
        var n = rowNames[i].innerHTML; 
        var g = rowGrades[i].innerHTML; 
        tempObj = {name: n, grade: g}; 
        list.push(tempObj); 
    }

    var templist = orderByGrade(list, gradeOrder); 
    gradeOrder = !gradeOrder;  
    
    for (var i = 0; i < templist.length; i++) {
        rowNames[i].innerHTML = templist[i].name; 
        rowGrades[i].innerHTML = templist[i].grade; 
    }
    
    colorRows(); 
}

function colorRows() {
    var elements = document.getElementsByClassName("rowEntry"); 
    var grades = document.getElementsByClassName("grade"); 
    var i; 
    for (i = 0; i < elements.length; i++) {
        if (grades[i].innerHTML < -3) {
            elements[i].setAttribute("class", "rowEntry table-danger"); 
        } else if (grades[i].innerHTML > 3) {
            elements[i].setAttribute("class", "rowEntry table-success"); 
        } else {
            elements[i].setAttribute("class", "rowEntry table-default"); 
        }
    }
    
}

function orderByName(list, nameList, order) {
    var tempList = []; 
    nameList.sort(); 
    for (var i = 0; i < list.length; i++) {
        var ind = 0; 
        var temp = list[i]; 
        for (var j = 0; j < list.length; j++) {
            if (temp.name == nameList[j]) {
                ind = j; 
                break; 
            }
        }
        tempList[ind] = temp; 
    }
    if (order === false) {
        tempList.reverse(); 
    }
    return tempList; 
}

function orderByGrade(list, order) {
    var tempList = []; 
    for (var i = -5; i <= 5; i++) {
        for (var j = 0; j < list.length; j++) {
            if (list[j].grade == i) {
                tempList.push(list[j])
            }
        }
    }
    if (order === false) {
        tempList.reverse(); 
    }
    return tempList; 
}