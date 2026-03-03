public class Vista extends JFrame{
    private MiPanel panel;

    public Vista(){
        setTitle("Ejercicio 4");
        setSize(600, 600);

        setLayout(new BorderLayout());
         panel = new MiPanel();
         add(panel, BorderLayout.CENTER);
    }

    public class MiPanel extends JPanel{
        int x1 = 0, x2 = 0;
        int y1 = 0, y2 = 0;
        private Color colorActual = Color.RED;
        private String forma = "LINEA";

        public void paintComponent(Graphics g){
            super.paintComponent(g);

            g.setColor(colorActual);

            if(forma.equals("LINEA")){
                g.drawLine(x1, y1, x2, y2);
            } else if(forma.equals("RECT")){
                g.drawRect(Math.min(x1,x2), Math.min(y1,y2), Math.abs(x2 - x1), Math.abs(y2-y1));
            }
        }

        public void setPos(int xini, int xfin, int yini, int yfin){
            x1 = xini;
            x2 = xfin;
            y1 = yini;
            y2 = yfin;
            repaint();
        }

        public void setTecla(char letra){
            switch(letra){
                case 'r': case 'R':
                    colorActual = Color.RED;
                    break;
                case 'g': case 'G':
                    colorActual = Color.GREEN;
                    break;
                case 'b': case 'B':
                    colorActual = Color.BLUE;
                    break;
                case 'l': case 'L':
                    forma = "LINEA";
                    break;
                case 'c': case 'C':
                    forma = "RECT";
                    break;
                default:
                    break;
            }
            repaint();

        }
    }

    public MiPanel getPanel(){
        return panel;
    }

    public void setPos(int xini, int xfin, int yini, int yfin){
        panel.setPos(xini, xfin, yini, yfin);
    }

    public void setTecla(char letra){
        panel.setTecla(letra);
    }
}