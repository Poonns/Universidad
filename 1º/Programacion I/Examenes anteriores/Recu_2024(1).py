""" Disponemos de una serie de matrices que almacenan informaci´on sobre las reservas de agua en todos los embalses
de Espa˜na. Cada matriz tiene dimensiones N × D, siendo N el n´umero de embalses de los que se tiene informaci´on,
y D el n´umero de d´ıas del a˜no (por simplicidad, consideraremos que es 365 en todos los casos). Cada celda [i][j]
de la matriz contiene un n´umero real entre 0 y 100 que indica el porcentaje de agua en el embalse i el d´ıa j.
Algunas celdas de la matriz contienen el valor None, lo que indica que el dispositivo de medida del nivel del agua
fall´o ese d´ıa.
Implementa en Python:
a) (1,5 puntos) Una funci´on, peor_embalse, que tenga como par´ametro una matriz como la descrita con la
informaci´on de un a˜no completo y devuelva como resultado el n´umero de embalse cuyo promedio anual de agua
embalsada sea m´as bajo. Para calcular el promedio, debes ignorar las celdas que contengan el valor None. """

def peor_embalse(matriz):
    peor_promedio = float('inf')
    indice_peor_embalse = -1
    
    for i in range(len(matriz)):
        suma_anual_embalse = 0
        dias_validos = 0
        for j in range(len(matriz[0])):
            if matriz[i][j] is not None:
                suma_anual_embalse += matriz[i][j]
                dias_validos += 1

        if dias_validos > 0:
            promedio = suma_anual_embalse / dias_validos

            if promedio < peor_promedio:
                peor_promedio = promedio
                indice_peor_embalse = i

    return indice_peor_embalse

""" b) (1,5 puntos) Una funci´on, d´ıa_mayor_aumento, que tenga como par´ametros dos matrices como las descritas
con la informaci´on de dos a˜nos completos (matriz1 y matriz2) y que devuelva como resultado el d´ıa (valor
de 0 a 364) para el que el promedio de agua almacenada en todos los embalses sufri´o un mayor incremento en
matriz2 con respecto a matriz1. Al igual que en la funci´on anterior, para calcular el promedio debes ignorar
las celdas que contengan el valor None. Si en ning´un d´ıa del a˜no hubo un incremento del agua embalsada con
respecto al a˜no anterior, la funci´on devolver´a None. Sup´on (sin comprobarlo) que las dos matrices tienen las
mismas dimensiones. """

def dia_mayor_aumento(matriz1, matriz2):
    mayor_incremento = float('-inf')
    indice_dia_mayor_incremento = -1

    for j in range(len(matriz1[0])):
        suma_matriz1 = 0
        suma_matriz2 = 0
        dias_validos = 0
        for i in range(len(matriz1)):
            if matriz1[i][j] is not None and matriz2[i][j] is not None:
                suma_matriz1 += matriz1[i][j]
                suma_matriz2 += matriz2[i][j]
                dias_validos += 1

        if dias_validos > 0:
            promedio1 = suma_matriz1 / dias_validos
            promedio2 = suma_matriz2 / dias_validos

            incremento = promedio2 - promedio1
            if incremento > mayor_incremento:
                mayor_incremento = incremento
                indice_dia_mayor_incremento = j

    return indice_dia_mayor_incremento if mayor_incremento > 0 else None

""" c) (2 puntos) Una funci´on, sistema_defectuoso, que tenga como par´ametros una matriz como la descrita con
la informaci´on de un a˜no completo y un valor T y que devuelva como resultado una lista con los embalses
(identificados por su n´umero de 0 a N − 1) para los que el sistema de medida fall´o (registr´o None) durante T o
m´as d´ıas consecutivos """

def sistema_defectuoso(matriz, T):
    sistema_fallido = []

    for i in range(len(matriz)):
        dias_fallos_seguidos = 0
        for j in range(len(matriz[0])):
            if matriz[i][j] is None:
                dias_fallos_seguidos += 1
            else:
                dias_fallos_seguidos = 0

            if dias_fallos_seguidos >= T:
                sistema_fallido.append(i)
                break

    return sistema_fallido