angulo = int(input('Introduce un angulo: '))

if angulo == 0:
    print('El angulo es nulo')
elif 0 < angulo < 90:
    print('El angulo es agudo')
elif angulo == 90:
    print('El angulo es recto')
else:
    print('El angulo es obtuso')