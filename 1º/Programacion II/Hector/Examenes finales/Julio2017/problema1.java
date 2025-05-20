package Hector.Examenes

import java.util.NoSuchElementException;

finales.Julio2017;

public class problema1 {
    class NotaEstudiante{
        private String dni;
        private int convocatoria;
        private Double nota;

        NotaEstudiante(String DNI, int convocatoria){
            this.dni = DNI;
            this.convocatoria = convocatoria;
        }

        String getDni(){
            return this.dni;
        }

        int getConvocatoria(){
            return this.convocatoria;
        }

        double getNota(){
            return nota;
        }

        void setNota(double nota){
            this.nota = nota;
        }

        boolean getPresentado(){
            return this.nota != null;
        }
    }

    class Acta{
        private String codigo;
        private int curso;
        private NotaEstudiante[] notas;

        public Acta(String codigo, int curso, NotaEstudiante[] notas){
            this.codigo = codigo;
            this.curso = curso;
            this.notas = notas;
        }

        void calificar(String DNI, double nota) throws NoSuchElementException{
            int indice = -1;
            int inicio = 0;
            int fin = notas.length -1;

            while(inicio <= fin){
                int media = (inicio + fin) / 2;
                int comparacion = notas[media].getDni().compareTo(DNI);

                if(comparacion == 0){
                    indice = media;
                    break;
                } else if(comparacion < 0){
                    inicio = media +1;
                } else {
                    fin = media -1;
                }
            }

            if(indice == -1){
                throw new NoSuchElementException("El Dni no está en la lista");
            } else{
                notas[indice].setNota(nota);
            }
        }
    }
}
