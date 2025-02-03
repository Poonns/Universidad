
""" La asociación valenciana de alergología dispone de una serie de captadores de polen situados en distintos puntos de
su territorio. La información recogida por un captador se almacena en una matriz, de modo que la celda [i][j]
contiene la cantidad de granos de polen detectados durante el día i para el tipo de polen j. Los nombres de
los distintos tipos de polen reconocidos por el captador se almacenan en una lista de cadenas, de modo que el
elemento [j] de la lista contiene el nombre del tipo de polen que se corresponde con la columna j de la matriz.
 """

tipos_polen = ['Olivo', 'Gramíneas', 'Ortigas', 'Castaño', 'Roble', 'Ciprés']

                # Oli Gra Ort Cas Rob Cip
cantidad_polen = [[13, 8, 22, 16, 10, 21], # día 0
                [ 2, 20, 21, 23, 3, 8], # día 1
                [ 3, 24, 6, 4, 20, 24], # día 2
                [12, 21, 22, 24, 6, 17], # día 3
                [11, 23, 24, 5, 18, 16]] # día 4

umbral = 12

dias_seguidos = 3

tipos_destacados = ['Gramíneas', 'Ortigas', 'Roble']

""" a) (1 punto) Una función, peor_día, que tenga como parámetro la matriz asociada a un captador y devuelva
como resultado el día en el que la suma total de granos de polen detectados (incluyendo todos los tipos de
polen) sea mayor. En el poco probable caso de que varios días proporcionen el mismo valor m´aximo, la funci´on
puede devolver cualquiera de ellos. Con los datos del ejemplo, la llamada """

def peor_dia(cantidad_polen):
    peor_dia = 0
    max_suma_elemento = 0
    for i in range(len(cantidad_polen)):
        suma_elementos_fila = 0
        for j in range(len(cantidad_polen[0])):
            suma_elementos_fila += cantidad_polen[i][j]
            if suma_elementos_fila > max_suma_elemento:
                max_suma_elemento = suma_elementos_fila
                peor_dia = i
    return peor_dia

""" b) (2 puntos) Una función, días_complejos, que tenga como parámetros la matriz asociada a un captador, la
lista con los nombres de los tipos de polen reconocidos por ese captador, un umbral(número entero) y otra lista
con los nombres de algunos tipos de polen destacados. La función debe devolver como resultado una lista con
los días en los que para cada uno de los tipos de polen destacados se registraron más granos de polen que los
indicados por el umbral dado. Lógicamente, todos los nombres en la lista de tipos de polen destacados aparecen
en la lista de tipos de polen reconocidos por el captador.  """

def dias_complejos(cantidad_polen, tipos_polen, umbral, tipos_destacados):
    lista = []
    for i in range(len(cantidad_polen)):
        cumple = True
        for tipo in tipos_destacados:
            for j in range(len(tipos_polen)):
                if tipos_polen[j] == tipo:
                    if cantidad_polen[i][j] < umbral:
                        cumple = False

        if cumple:
            lista.append(i)

    return lista
         
""" c) (2 puntos) Una función, alerta_por_continuidad, que tenga como parámetros la matriz asociada a un
captador, la lista con los nombres de los tipos de polen reconocidos por ese captador, un umbral y una cantidad
de días n. La función debe devolver como resultado una lista con los nombres de los tipos de polen para los
que se han registrado durante n o más días consecutivos una cantidad de granos de polen superior al umbral
dado. Con los datos del ejemplo, la llamada """

def alerta_por_comtinuidad(cantidad_polen, tipos_polen, umbral, dias_seguidos):
    lista = []
    for j in range(len(tipos_polen)):
        tipo = tipos_polen[j]
        contador = 0

        for i in range(len(cantidad_polen)):
            if cantidad_polen[i][j] >= umbral:
                contador += 1
            else:
                contador = 0

        if contador >= dias_seguidos:
            lista.append(tipo)

    return lista

