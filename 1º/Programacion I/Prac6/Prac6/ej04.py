def evaluacion_test(plantilla, respuesta):
    aciertos = 0
    fallos = 0
    votos_en_blanco = 0

    for i in range(min(len(respuesta), numero_total_de_preguntas)):
        if respuesta[i] == correcion[i]:
            aciertos += 1
        elif respuesta[i] == '*':
            votos_en_blanco += 1
        else:
            fallos += 1
    
    return [aciertos, fallos, votos_en_blanco]


name_file = input('Introduce el nombre del archivo: ')

file = open(name_file, 'r')
lineas = file.readlines()

correcion = lineas[0]
numero_total_de_preguntas = len(correcion)

escribir_archivo = input('Introduce el nombre del archivo donde deseas escribir las notas: ')
archivo = open(escribir_archivo, 'w')

for linea in lineas[1:]:
    dni, respuesta = linea.split('#')

    aciertos, fallos, votos_en_blanco = evaluacion_test(correcion, respuesta)
    nota = 10 * (aciertos - fallos) / numero_total_de_preguntas 

    if nota < 0:
        nota = 0
       
    if nota < 5:
        archivo.write(f'DNI: {dni} NOTA: NO APTO\n')
    else:
        archivo.write(f'DNI: {dni} NOTA: APTO ({nota: .2f})\n')

archivo.close()
file.close()