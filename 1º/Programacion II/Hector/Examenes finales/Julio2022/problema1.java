package Hector.Examenes finales.Julio2022;

public class problema1 {
    // Clase Libro
    public class Libro {
        private String ISBN;
        private double puntos; // promedio de puntuación
        private int votos;     // número de compradores que puntuaron

        public Libro(String ISBN, double puntos, int votos) {
            this.ISBN = ISBN;
            this.puntos = puntos;
            this.votos = votos;
        }

        public String getISBN() {
            return ISBN;
        }

        public double getPuntos() {
            return puntos;
        }

        public int getVotos() {
            return votos;
        }
    }

    // Clase Materia
    public class Materia {
        private static final int TAM_INICIAL = 10;
        private Libro[] libros; // vector de libros ordenados por puntuación descendente
        private int cantidad;   // cantidad real de libros almacenados

        public Materia() {
            libros = new Libro[TAM_INICIAL];
            cantidad = 0;
        }

        void eliminar(int minVotos){
            int contador = 0;
            for(int i = 0; i < cantidad; i++){
                if(libros[i].getVotos() < minVotos){
                    contador++;
                }
            }

            Libro[] nuevo = new Libro[libros.length];
            int j = 0;
            for(int i = 0; i < cantidad; i++){
                if(libros[i].getVotos() >= minVotos){
                    nuevo[j++] = libros[i];
                }
            }

            libros = nuevo;
            cantidad = j;
        }

        Libro getLibro(String ISBN){
            for(int i = 0; i < cantidad; i++){
                if(libros[i].getISBN().equals(ISBN)){
                    return libros[i];
                }
            }
            return null;
        }

        String[] mejoresLibros(Materia otraMateria, int n){
            String[] nuevo = new String[n];
            int i = 0;
            int j = 0;
            int k = 0;

            while(i < cantidad && j < otraMateria.cantidad && k < n){
                if(libros[i].getPuntos() > otraMateria.libros[j].getPuntos()){
                    nuevo[k++] = libros[i++].getISBN();
                } else {
                    nuevo[k++] = otraMateria.libros[j++].getISBN();
                }
            }

            while(i < cantidad && k < n) {
                nuevo[k++] = libros[i].getISBN();
            }

            while(j < otraMateria.cantidad && k < n) {
                nuevo[k++] = otraMateria.libros[j].getISBN();
            }

            if(k < n){
                String[] vector = new String[k];
                for(int h = 0; h < k; h++){
                    vector[h] = nuevo[h];
                }
                return vector;
            }
            return nuevo;
        }
    }
}
