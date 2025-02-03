""" Una autoescuela guarda la informaci´on acerca de las clases pr´acticas de sus estudiantes en una matriz como esta:
 """
matriz = [[None, None, None, None, None, None, 1, 1, 2, 1],
[None, None, 2, 0.5, 1, 1, 0, 0, 1, 0.5],
[None, 1, 0, 0.5, 1, 1, 2, 2.5, 1, None],
[None, 2, 2, 0.5, 2, 3, None, None, None, None],
[ 1, 1, 0, 0, 2, 1, 1, 0, 1, 1],
[None, 2, 2, 0.5, 2, 3, None, None, None, None]]

""" Cada fila se corresponde con un estudiante diferente y cada columna, con un d´ıa determinado dentro de un intervalo
de tiempo. Los nombres de los estudiantes se guardan en una lista como la mostrada a continuaci´on, conservando
el mismo orden en el que aparecen en la matriz: """

estudiantes = ['Martina', 'Noa', 'Iria', 'Angela', 'Alvaro', 'Paloma']

""" Por tanto, la fila 0 de la matriz se corresponde con Martina, la fila 1 con Noa, y as´ı sucesivamente.
Si en la posici´on [i][j] de la matriz encontramos un n´umero, este representa el n´umero de horas de pr´acticas
con instructor que el estudiante i ha realizado el d´ıa j. No todos los estudiantes empiezan el primer d´ıa: aquellos
que empiezan m´as tarde, tienen None los d´ıas anteriores al de su primera clase. Adem´as, aquellos estudiantes que
se examinan y obtienen el carnet de conducir, tienen None el d´ıa del examen y los sucesivos.
Por ejemplo, Martina empieza sus pr´acticas el d´ıa 6 (contamos los d´ıas desde 0) y Noa, el d´ıa 2, aunque ninguna
ha terminado. Iria ha acabado sus pr´acticas el d´ıa 9 y ´Angela, el d´ıa 6.
Implementa en Python:

a) (1 punto) Una funci´on, carnets_por_d´ıa, que tenga como par´ametros una matriz como la del ejemplo y un
n´umero de d´ıa (recuerda que empezamos a contar los d´ıas desde 0). La funci´on debe devolver la cantidad
de estudiantes que han obtenido su carnet ese d´ıa. El primer d´ıa ning´un estudiante puede haber obtenido su
carnet  """

def carnet_por_dia(matriz, dia):
    resultados = []
    for j in range(len(matriz[0])-1, 1, -1):
            contador_aprobados = 0
            for i in range(len(matriz)):
                if matriz[i][j] is None and matriz[i][j-1] is not None:
                        contador += 1
                resultados.append((j, contador_aprobados))
    return resultados

""" b) (1 punto) Una funci´on, horas_por_estudiante, que tenga como par´ametros una matriz como la del ejemplo y
un n´umero de estudiante. La funci´on debe devolver la cantidad total de horas de pr´acticas con instructor que
ese estudiante ha realizado en el periodo de tiempo representado por la matriz.
Considerando los datos del ejemplo, para el estudiante n´umero 0 obtendr´ıamos 5 horas y para el estudiante
n´umero 1, 6 horas """

def horas_por_estudiante(matriz, num_estudiante):
    tiempo_alumno = 0
    for j in range(len(matriz[0])):
        if matriz[num_estudiante][j] is not None:
            tiempo_alumno += matriz[num_estudiante][j]
            
    return tiempo_alumno

""" c) (2 puntos) Una funci´on, mejor_estudiante, que tenga como par´ametros una matriz como la del ejemplo y una
lista de nombres de estudiantes. La funci´on debe devolver el nombre del estudiante que, habiendo acabado su
instrucci´on (es decir, que se ha examinado y obtenido el carnet) haya necesitado menos horas de clase. En caso
de empate entre varios estudiantes, la funci´on puede devolver cualquiera de ellos. Esta funci´on debe llamar
obligatoriamente a la funci´on definida en el apartado anterior.
Considerando los datos del ejemplo, la funci´on indicar´ıa que Iria es la mejor estudiante, pues es quien ha acabado
la instrucci´on con el menor n´umero de horas (9 horas). F´ıjate en que, por ejemplo, Martina ha empleado menos
horas (5 horas), pero no ha acabado su instrucci´on, por lo que no se tiene en cuenta. """

def mejor_estudiante(matriz, estudiantes):
    mejor_tiempo = float('inf')
    mejor_alumno = None

    for i in range(len(estudiantes)):
        if matriz[i][-1] is None:
            tiempo_alumno = horas_por_estudiante(matriz, i)

            if tiempo_alumno < mejor_tiempo:
                mejor_tiempo = tiempo_alumno
                mejor_alumno = estudiantes[i]

    return mejor_alumno

    